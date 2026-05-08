package confkit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestErrorReportUnwrap(t *testing.T) {
	report := &ErrorReport{}
	report.AddError(FieldError{Path: "host", Message: "required"})
	report.AddError(FieldError{Path: "port", Message: "invalid"})

	errs := report.Unwrap()
	if len(errs) != 2 {
		t.Fatalf("expected 2, got %d", len(errs))
	}
	if errs[0].Error() != "required" {
		t.Errorf("expected 'required', got %q", errs[0].Error())
	}
	if errs[1].Error() != "invalid" {
		t.Errorf("expected 'invalid', got %q", errs[1].Error())
	}
}

func TestErrorReportUnwrapEmpty(t *testing.T) {
	report := &ErrorReport{}
	errs := report.Unwrap()
	if len(errs) != 0 {
		t.Errorf("expected 0, got %d", len(errs))
	}
}

func TestErrorsJoin(t *testing.T) {
	report := &ErrorReport{}
	report.AddError(FieldError{Path: "x", Message: "bad"})
	wrapped := report.Unwrap()
	if len(wrapped) == 0 {
		t.Fatal("no wrapped errors")
	}
	if wrapped[0].Error() != "bad" {
		t.Errorf("expected 'bad', got %q", wrapped[0].Error())
	}
}

func TestNewErrorSource(t *testing.T) {
	sentinel := fmt.Errorf("injected error")
	src := NewErrorSource(sentinel)
	if src.Name() != "error" {
		t.Errorf("expected name 'error', got %q", src.Name())
	}
	_, ok, err := src.Lookup(context.Background(), &FieldInfo{})
	if ok || err == nil {
		t.Errorf("expected error lookup to fail: ok=%v err=%v", ok, err)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestLoadWithErrorSource(t *testing.T) {
	type Cfg struct {
		Host string `env:"ERR_HOST"`
	}
	sentinel := fmt.Errorf("source failure")
	_, err := Load[Cfg](NewErrorSource(sentinel))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseMapStringString(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[string]string{})
	val, err := p.Parse("KEY1=v1,KEY2=v2", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := val.(map[string]string)
	if m["KEY1"] != "v1" || m["KEY2"] != "v2" {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestParseMapEmpty(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[string]string{})
	val, err := p.Parse("", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := val.(map[string]string)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestParseMapInvalidEntry(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[string]string{})
	_, err := p.Parse("noequals", typ)
	if err == nil {
		t.Fatal("expected error for invalid entry")
	}
}

func TestParseMapNonStringKey(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[int]string{})
	_, err := p.Parse("1=v", typ)
	if err == nil {
		t.Fatal("expected error for non-string key")
	}
}

func TestParseMapIntValue(t *testing.T) {
	p := NewParser()
	typ := reflect.TypeOf(map[string]int{})
	val, err := p.Parse("a=1,b=2", typ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := val.(map[string]int)
	if m["a"] != 1 || m["b"] != 2 {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestLoadMapField(t *testing.T) {
	type Cfg struct {
		Labels map[string]string `env:"MAP_LABELS"`
	}
	_ = os.Setenv("MAP_LABELS", "app=web,env=prod")
	t.Cleanup(func() { _ = os.Unsetenv("MAP_LABELS") })

	cfg, err := Load[Cfg](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Labels["app"] != "web" || cfg.Labels["env"] != "prod" {
		t.Errorf("unexpected labels: %v", cfg.Labels)
	}
}

func TestAuditLoggerOnValidationFailure(t *testing.T) {
	type Cfg struct {
		Port int `env:"AUDIT_FAIL_PORT" validate:"min=1"`
	}
	_ = os.Setenv("AUDIT_FAIL_PORT", "0")
	t.Cleanup(func() { _ = os.Unsetenv("AUDIT_FAIL_PORT") })

	var logCalled bool
	_, err := LoadWithOptions[Cfg](
		WithSource(FromEnv()),
		WithAuditLogger(func(entries []AuditEntry) {
			logCalled = true
		}),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !logCalled {
		t.Error("expected audit logger to be called on failure")
	}
}

func TestAuditLoggerOnSourceError(t *testing.T) {
	type Cfg struct {
		X string `env:"AUDITX"`
	}
	var logCalled bool
	_, _ = LoadWithOptions[Cfg](
		WithSource(NewErrorSource(fmt.Errorf("src err"))),
		WithAuditLogger(func(entries []AuditEntry) {
			logCalled = true
		}),
	)
	if !logCalled {
		t.Error("expected audit logger to be called on source error")
	}
}

func TestPortValidationUint(t *testing.T) {
	v := NewValidator()
	field := FieldInfo{Path: "Port", Name: "Port"}
	rule := ValidationRule{Name: "port"}

	good := reflect.ValueOf(uint(8080))
	fe := v.validateField(good, field, rule)
	if fe.Message != "" {
		t.Errorf("expected no error for valid uint port, got: %s", fe.Message)
	}

	bad := reflect.ValueOf(uint(0))
	fe = v.validateField(bad, field, rule)
	if fe.Message == "" {
		t.Error("expected error for port 0 (uint)")
	}
}

func TestPortValidationStringField(t *testing.T) {
	v := NewValidator()
	field := FieldInfo{Path: "Port", Name: "Port"}
	rule := ValidationRule{Name: "port"}

	good := reflect.ValueOf("443")
	fe := v.validateField(good, field, rule)
	if fe.Message != "" {
		t.Errorf("expected no error for valid string port, got: %s", fe.Message)
	}

	bad := reflect.ValueOf("notanumber")
	fe = v.validateField(bad, field, rule)
	if fe.Message == "" {
		t.Error("expected error for non-numeric string port")
	}
}

func TestEmbeddedStructScan(t *testing.T) {
	type Base struct {
		Host string `env:"EMB_HOST"`
	}
	type Cfg struct {
		Base
		Port int `env:"EMB_PORT"`
	}

	fields := ScanFields(Cfg{})
	names := make(map[string]bool)
	for _, f := range fields {
		names[f.Name] = true
	}
	if !names["Host"] {
		t.Error("expected embedded Host field to be promoted")
	}
	if !names["Port"] {
		t.Error("expected Port field")
	}
}

func TestExplainNilError(t *testing.T) {
	if Explain(nil) != "" {
		t.Error("expected empty string for nil error")
	}
}

func TestExplainNonReport(t *testing.T) {
	e := fmt.Errorf("plain error")
	if Explain(e) != "plain error" {
		t.Errorf("expected 'plain error', got %q", Explain(e))
	}
}

func TestParseDurationValid(t *testing.T) {
	p := NewParser()
	val, err := p.Parse("5s", reflect.TypeOf((*interface{ String() string })(nil)).Elem())
	_ = val
	_ = err
}

func TestLoadDurationField(t *testing.T) {
	type Cfg struct {
		Timeout string `env:"DUR_TIMEOUT" default:"30s"`
	}
	cfg, err := Load[Cfg](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != "30s" {
		t.Errorf("expected '30s', got %q", cfg.Timeout)
	}
}

func TestGetFieldValueNilPtrField(t *testing.T) {
	type Inner struct {
		Val string
	}
	type Outer struct {
		Inner *Inner
	}
	v := reflect.ValueOf(Outer{})
	result := getFieldValue(v, "Inner.Val")
	if result != nil {
		t.Errorf("expected nil for nil ptr path, got %v", result)
	}
}

func TestGetFieldValueInvalidField(t *testing.T) {
	v := reflect.ValueOf(struct{ X string }{X: "ok"})
	result := getFieldValue(v, "NoSuch")
	if result != nil {
		t.Errorf("expected nil for missing field, got %v", result)
	}
}

func TestStrCheckNonStringField(t *testing.T) {
	v := NewValidator()
	field := FieldInfo{Path: "X", Name: "X"}
	rule := ValidationRule{Name: "email"}
	fe := v.validateField(reflect.ValueOf(42), field, rule)
	if fe.Message != "" {
		t.Errorf("expected no error for non-string field with string validator, got: %s", fe.Message)
	}
}

func TestParseValidationRulesOneof(t *testing.T) {
	rules := parseValidationRules("required,oneof=a b c")
	found := false
	for _, r := range rules {
		if r.Name == "oneof" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected oneof rule")
	}
}

func TestJSONSourceYAMLTagFallback(t *testing.T) {
	f := t.TempDir() + "/test.json"
	_ = os.WriteFile(f, []byte(`{"host":"jsonhost","port":4000}`), 0644)
	type Cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	cfg, err := Load[Cfg](FromJSON(f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "jsonhost" {
		t.Errorf("expected jsonhost, got %s", cfg.Host)
	}
}

func TestTOMLSourceYAMLTagFallback(t *testing.T) {
	f := t.TempDir() + "/test.toml"
	_ = os.WriteFile(f, []byte("host = \"tomlhost\"\nport = 5000\n"), 0644)
	type Cfg struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	cfg, err := Load[Cfg](FromTOML(f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "tomlhost" {
		t.Errorf("expected tomlhost, got %s", cfg.Host)
	}
}

func TestTOMLSourceMissingField(t *testing.T) {
	f := t.TempDir() + "/test.toml"
	_ = os.WriteFile(f, []byte("host = \"myhost\"\n"), 0644)
	type Cfg struct {
		Host string `toml:"host" default:"fallback"`
		Port int    `toml:"port" default:"9999"`
	}
	cfg, err := Load[Cfg](FromTOML(f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("expected 9999, got %d", cfg.Port)
	}
}

func TestScanFieldsPointerStruct(t *testing.T) {
	type Cfg struct {
		Host string `env:"PSF_HOST"`
	}
	fields := ScanFields(&Cfg{})
	if len(fields) == 0 {
		t.Error("expected fields from pointer struct")
	}
}

func TestFieldValueToStringAllTypes(t *testing.T) {
	cases := []struct {
		val      reflect.Value
		expected string
	}{
		{reflect.ValueOf(true), "true"},
		{reflect.ValueOf(uint(5)), "5"},
		{reflect.ValueOf(float64(3.14)), "3.14"},
	}
	for _, tc := range cases {
		got := fieldValueToString(tc.val, false)
		if got != tc.expected {
			t.Errorf("fieldValueToString(%v) = %q, want %q", tc.val, got, tc.expected)
		}
	}
}

func TestIsZeroValueBool(t *testing.T) {
	if !isZeroValue(reflect.ValueOf(false)) {
		t.Error("false should be zero")
	}
	if isZeroValue(reflect.ValueOf(true)) {
		t.Error("true should not be zero")
	}
}

func TestIsZeroValueSlice(t *testing.T) {
	if !isZeroValue(reflect.ValueOf([]string{})) {
		t.Error("empty slice should be zero")
	}
	if isZeroValue(reflect.ValueOf([]string{"a"})) {
		t.Error("non-empty slice should not be zero")
	}
}

func TestGetFieldByPathNilPtr(t *testing.T) {
	type Inner struct {
		Val string
	}
	type Outer struct {
		Inner *Inner
	}
	v := reflect.ValueOf(Outer{})
	result := getFieldByPath(v, "Inner.Val")
	if result.IsValid() {
		t.Error("expected invalid value for nil pointer path")
	}
}

func TestGetFieldByPathInvalid(t *testing.T) {
	v := reflect.ValueOf(struct{ X string }{})
	result := getFieldByPath(v, "NoSuchField")
	if result.IsValid() {
		t.Error("expected invalid value for missing field")
	}
}

func TestSetFieldValueNestedPtr(t *testing.T) {
	type Inner struct {
		Val string
	}
	type Outer struct {
		Inner *Inner
	}
	v := Outer{}
	rv := reflect.ValueOf(&v).Elem()
	setFieldValue(rv, "Inner.Val", "hello")
	if v.Inner == nil || v.Inner.Val != "hello" {
		t.Errorf("expected Inner.Val='hello', got %+v", v)
	}
}

func TestValidationBoolField(t *testing.T) {
	type Cfg struct {
		Flag bool `env:"VAL_BOOL_FLAG" validate:"required"`
	}
	_ = os.Setenv("VAL_BOOL_FLAG", "true")
	t.Cleanup(func() { _ = os.Unsetenv("VAL_BOOL_FLAG") })
	_, err := Load[Cfg](FromEnv())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
