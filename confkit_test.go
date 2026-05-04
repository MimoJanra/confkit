package confkit

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestScanFields(t *testing.T) {
	type testConfig struct {
		Port int    `env:"PORT" default:"8080"`
		Mode string `env:"MODE" default:"dev"`
	}

	config := testConfig{}
	fields := ScanFields(config)

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	// Check Port field
	portField := fields[0]
	if portField.Name != "Port" {
		t.Errorf("expected field name Port, got %s", portField.Name)
	}
	if portField.Tags["env"] != "PORT" {
		t.Errorf("expected env tag PORT, got %s", portField.Tags["env"])
	}
	if portField.Tags["default"] != "8080" {
		t.Errorf("expected default tag 8080, got %s", portField.Tags["default"])
	}

	// Check Mode field
	modeField := fields[1]
	if modeField.Name != "Mode" {
		t.Errorf("expected field name Mode, got %s", modeField.Name)
	}
}

func TestParseString(t *testing.T) {
	p := NewParser()
	result, err := p.Parse("hello", reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestParseEmpty(t *testing.T) {
	p := NewParser()

	// Empty string should return zero value for each type
	types := []reflect.Type{
		reflect.TypeOf(""),
		reflect.TypeOf(0),
		reflect.TypeOf(false),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf([]string{}),
	}

	for _, typ := range types {
		result, err := p.Parse("", typ)
		if err != nil {
			t.Errorf("Parse('', %v): unexpected error: %v", typ, err)
		}
		_ = result
	}
}

func TestParseUnsupportedType(t *testing.T) {
	p := NewParser()
	// map is not supported
	_, err := p.Parse("test", reflect.TypeOf(map[string]string{}))
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"invalid", false, true},
	}

	p := NewParser()
	for _, tc := range tests {
		result, err := p.Parse(tc.input, reflect.TypeOf(false))
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q): got error %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if err == nil && result != tc.expected {
			t.Errorf("Parse(%q): expected %v, got %v", tc.input, tc.expected, result)
		}
	}
}

func TestParseInt(t *testing.T) {
	p := NewParser()

	tests := []struct {
		input    string
		typ      reflect.Type
		expected any
		wantErr  bool
	}{
		{"42", reflect.TypeOf(int(0)), int(42), false},
		{"255", reflect.TypeOf(uint8(0)), uint8(255), false},
		{"256", reflect.TypeOf(uint8(0)), nil, true}, // overflow
		{"-128", reflect.TypeOf(int8(0)), int8(-128), false},
		{"100", reflect.TypeOf(int32(0)), int32(100), false},
		{"200", reflect.TypeOf(int16(0)), int16(200), false},
		{"400", reflect.TypeOf(int64(0)), int64(400), false},
		{"invalid", reflect.TypeOf(int(0)), nil, true},
		{"32768", reflect.TypeOf(int16(0)), nil, true}, // overflow int16
	}

	for _, tc := range tests {
		result, err := p.Parse(tc.input, tc.typ)
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q): got error %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if err == nil && result != tc.expected {
			t.Errorf("Parse(%q): expected %v, got %v", tc.input, tc.expected, result)
		}
	}
}

func TestParseFloat(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := result.(float64)
	if f < 3.13 || f > 3.15 {
		t.Errorf("expected ~3.14, got %v", f)
	}
}

func TestParseDuration(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("5s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := result.(time.Duration)
	if d != 5*time.Second {
		t.Errorf("expected 5s, got %v", d)
	}
}

func TestParseTime(t *testing.T) {
	p := NewParser()

	result, err := p.Parse("2026-01-01T00:00:00Z", reflect.TypeOf(time.Time{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tm := result.(time.Time)
	if tm.Year() != 2026 || tm.Month() != time.January || tm.Day() != 1 {
		t.Errorf("unexpected date: %v", tm)
	}
}

func TestParseSlice(t *testing.T) {
	p := NewParser()

	// []string
	result, err := p.Parse("a,b,c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice := result.([]string)
	if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
		t.Errorf("expected [a b c], got %v", slice)
	}

	// []int
	result, err = p.Parse("1,2,3", reflect.TypeOf([]int{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intSlice := result.([]int)
	if len(intSlice) != 3 || intSlice[0] != 1 || intSlice[1] != 2 || intSlice[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", intSlice)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	type Config struct {
		Port int    `env:"DEFAULT_PORT" default:"8080"`
		Mode string `env:"DEFAULT_MODE" default:"dev"`
	}

	// Don't set env vars, so defaults apply

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Port)
	}
	if cfg.Mode != "dev" {
		t.Errorf("expected Mode 'dev', got %q", cfg.Mode)
	}
}

func TestLoadFromEnv(t *testing.T) {
	type Config struct {
		Port int    `env:"TEST_PORT"`
		Host string `env:"TEST_HOST"`
	}

	t.Setenv("TEST_PORT", "3000")
	t.Setenv("TEST_HOST", "localhost")

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %q", cfg.Host)
	}
}

func TestLoadErrorReportParsing(t *testing.T) {
	type Config struct {
		Port int `env:"BAD_PORT"`
	}

	t.Setenv("BAD_PORT", "not-a-number")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	fieldErr := report.Errors[0]
	if fieldErr.Kind != ErrorKindParse {
		t.Errorf("expected ErrorKindParse, got %s", fieldErr.Kind)
	}
	if fieldErr.Path != "Port" {
		t.Errorf("expected Path 'Port', got %q", fieldErr.Path)
	}
}

func TestExplainError(t *testing.T) {
	type Config struct {
		API string `env:"API_KEY" secret:"true"`
	}

	t.Setenv("API_KEY", "super-secret-key")

	_, err := Load[Config](FromEnv())
	if err == nil {
		// Manually create an error to test
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "API",
			Source:  "env API_KEY",
			Kind:    ErrorKindParse,
			Message: "invalid value",
			Value:   "super-secret-key",
			Secret:  true,
		})
		err = report
	}

	explained := Explain(err)
	if !strings.Contains(explained, "<redacted>") {
		t.Errorf("expected secret to be redacted, got: %s", explained)
	}
}

func TestLoadFromYAML(t *testing.T) {
	type Config struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
		Mode string `yaml:"mode"`
	}

	cfg, err := Load[Config](FromYAML("testdata/config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %q", cfg.Host)
	}
	if cfg.Mode != "production" {
		t.Errorf("expected Mode 'production', got %q", cfg.Mode)
	}
}

func TestLoadFromJSON(t *testing.T) {
	type Config struct {
		Port int    `json:"port"`
		Host string `json:"host"`
		Mode string `json:"mode"`
	}

	cfg, err := Load[Config](FromJSON("testdata/config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 2000 {
		t.Errorf("expected Port 2000, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Host)
	}
	if cfg.Mode != "development" {
		t.Errorf("expected Mode 'development', got %q", cfg.Mode)
	}
}

func TestLoadMultipleSources(t *testing.T) {
	type Config struct {
		Port int    `yaml:"port" env:"PORT" default:"8080"`
		Host string `yaml:"host" env:"HOST" default:"localhost"`
		Mode string `yaml:"mode" env:"MODE" default:"dev"`
	}

	t.Setenv("MODE", "test")

	cfg, err := Load[Config](
		FromYAML("testdata/config.yaml"),
		FromEnv(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// YAML provides port and host, env overrides mode
	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000 from YAML, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost' from YAML, got %q", cfg.Host)
	}
	if cfg.Mode != "test" {
		t.Errorf("expected Mode 'test' from env override, got %q", cfg.Mode)
	}
}

func TestValidationRequired(t *testing.T) {
	type Config struct {
		Host string `env:"TEST_REQUIRED_HOST" validate:"required"`
	}

	// Don't set the env var, so it will be empty
	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for required field, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	fieldErr := report.Errors[0]
	if fieldErr.Kind != ErrorKindValidation {
		t.Errorf("expected ErrorKindValidation, got %s", fieldErr.Kind)
	}
	if fieldErr.Path != "Host" {
		t.Errorf("expected Path 'Host', got %q", fieldErr.Path)
	}
}

func TestValidationMin(t *testing.T) {
	type Config struct {
		Port int `env:"TEST_MIN_PORT" validate:"min=1"`
	}

	t.Setenv("TEST_MIN_PORT", "0")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for min constraint, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	fieldErr := report.Errors[0]
	if fieldErr.Kind != ErrorKindValidation {
		t.Errorf("expected ErrorKindValidation, got %s", fieldErr.Kind)
	}
}

func TestValidationMax(t *testing.T) {
	type Config struct {
		Port int `env:"TEST_MAX_PORT" validate:"max=65535"`
	}

	t.Setenv("TEST_MAX_PORT", "99999")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for max constraint, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}
}

func TestValidationMultipleRules(t *testing.T) {
	type Config struct {
		Port int `env:"TEST_MULTI_PORT" validate:"required,min=1,max=65535"`
	}

	t.Setenv("TEST_MULTI_PORT", "8080")

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Port)
	}
}

func TestNestedStructsYAML(t *testing.T) {
	type DatabaseConfig struct {
		URL      string `yaml:"url"`
		PoolSize int    `yaml:"pool_size"`
	}

	type Config struct {
		Port     int            `yaml:"port"`
		Database DatabaseConfig `yaml:"database"`
	}

	cfg, err := Load[Config](FromYAML("testdata/config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", cfg.Port)
	}

	if cfg.Database.URL != "postgres://localhost:5432/app" {
		t.Errorf("expected Database.URL 'postgres://localhost:5432/app', got %q", cfg.Database.URL)
	}

	if cfg.Database.PoolSize != 20 {
		t.Errorf("expected Database.PoolSize 20, got %d", cfg.Database.PoolSize)
	}
}

func TestNestedStructsJSON(t *testing.T) {
	type DatabaseConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	type Config struct {
		Mode     string         `json:"mode"`
		Database DatabaseConfig `json:"database"`
	}

	cfg, err := Load[Config](FromJSON("testdata/config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != "development" {
		t.Errorf("expected Mode 'development', got %q", cfg.Mode)
	}

	if cfg.Database.Host != "0.0.0.0" {
		t.Errorf("expected Database.Host '0.0.0.0', got %q", cfg.Database.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port 5432, got %d", cfg.Database.Port)
	}
}

func TestLoadFromTOML(t *testing.T) {
	type Config struct {
		Port int    `toml:"port"`
		Host string `toml:"host"`
	}

	cfg, err := Load[Config](FromTOML("testdata/config.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected Port 9000, got %d", cfg.Port)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Host)
	}
}

func TestNestedStructsTOML(t *testing.T) {
	type DatabaseConfig struct {
		Host     string        `toml:"host"`
		Port     int           `toml:"port"`
		Username string        `toml:"username"`
		Password string        `toml:"password" secret:"true"`
		Timeout  time.Duration `toml:"timeout"`
	}

	type Config struct {
		Port     int            `toml:"port"`
		Host     string         `toml:"host"`
		Database DatabaseConfig `toml:"database"`
	}

	cfg, err := Load[Config](FromTOML("testdata/config.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected Port 9000, got %d", cfg.Port)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Host)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected Database.Host 'db.example.com', got %q", cfg.Database.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port 5432, got %d", cfg.Database.Port)
	}

	if cfg.Database.Username != "admin" {
		t.Errorf("expected Database.Username 'admin', got %q", cfg.Database.Username)
	}

	if cfg.Database.Password != "secret123" {
		t.Errorf("expected Database.Password 'secret123', got %q", cfg.Database.Password)
	}

	if cfg.Database.Timeout != 30*time.Second {
		t.Errorf("expected Database.Timeout 30s, got %v", cfg.Database.Timeout)
	}
}

func TestNestedStructsValidation(t *testing.T) {
	type DatabaseConfig struct {
		Host string `yaml:"host" validate:"required"`
		Port int    `yaml:"port" validate:"min=1,max=65535"`
	}

	type Config struct {
		Port     int            `yaml:"port" default:"8080"`
		Database DatabaseConfig `yaml:"database"`
	}

	cfg, err := Load[Config](FromYAML("testdata/config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port 5432, got %d", cfg.Database.Port)
	}
}

func TestValidationWithSecretRedaction(t *testing.T) {
	type Config struct {
		APIKey string `env:"TEST_SECRET" validate:"required" secret:"true"`
	}

	// Manually create a validation error to test redaction
	report := &ErrorReport{}
	report.AddError(FieldError{
		Path:    "APIKey",
		Kind:    ErrorKindValidation,
		Message: "invalid format",
		Value:   "super-secret-12345",
		Secret:  true,
	})

	formatted := report.Format()
	if !strings.Contains(formatted, "<redacted>") {
		t.Errorf("expected secret to be redacted in error, got: %s", formatted)
	}
	if strings.Contains(formatted, "super-secret-12345") {
		t.Errorf("secret value leaked in error: %s", formatted)
	}
}

func TestCustomValidator(t *testing.T) {
	// Register a custom validator for testing
	RegisterValidator("positive", func(v reflect.Value) error {
		if v.Kind() == reflect.Int && v.Int() <= 0 {
			return fmt.Errorf("must be positive")
		}
		return nil
	})

	type Config struct {
		Count int `env:"TEST_COUNT" validate:"positive"`
	}

	t.Setenv("TEST_COUNT", "-5")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for custom validator, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	fieldErr := report.Errors[0]
	if fieldErr.Rule != "positive" {
		t.Errorf("expected rule 'positive', got %s", fieldErr.Rule)
	}
	if !strings.Contains(fieldErr.Message, "must be positive") {
		t.Errorf("expected 'must be positive' in error message, got: %s", fieldErr.Message)
	}
}

func TestValidationOneOf(t *testing.T) {
	type Config struct {
		LogLevel string `env:"TEST_LOGLEVEL" validate:"oneof=debug,info,warn,error"`
	}

	t.Setenv("TEST_LOGLEVEL", "invalid")

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for oneof, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}

	if report.Errors[0].Rule != "oneof" {
		t.Errorf("expected rule 'oneof', got %s", report.Errors[0].Rule)
	}
}

func TestValidationOneOfValid(t *testing.T) {
	type Config struct {
		LogLevel string `env:"TEST_LOGLEVEL_VALID" validate:"oneof=debug,info,warn,error"`
	}

	t.Setenv("TEST_LOGLEVEL_VALID", "info")

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel 'info', got %q", cfg.LogLevel)
	}
}

func TestValidationOneOfWithRequired(t *testing.T) {
	type Config struct {
		Level string `env:"TEST_LEVEL_REQUIRED" validate:"required,oneof=debug,info"`
	}

	// Don't set env var — required should fail
	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if report.Errors[0].Rule != "required" {
		t.Errorf("expected rule 'required', got %s", report.Errors[0].Rule)
	}
}

func TestValidationStringMinMax(t *testing.T) {
	type Config struct {
		Name string `env:"TEST_NAME_LEN" validate:"min=3,max=10"`
	}

	t.Setenv("TEST_NAME_LEN", "ab")
	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for min string length, got nil")
	}

	t.Setenv("TEST_NAME_LEN", "this-is-too-long-for-max")
	_, err = Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for max string length, got nil")
	}

	t.Setenv("TEST_NAME_LEN", "hello")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "hello" {
		t.Errorf("expected Name 'hello', got %q", cfg.Name)
	}
}

func TestValidationFloatMinMax(t *testing.T) {
	type Config struct {
		Ratio float64 `env:"TEST_RATIO" validate:"min=0,max=1"`
	}

	t.Setenv("TEST_RATIO", "-0.5")
	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for float min, got nil")
	}

	t.Setenv("TEST_RATIO", "1.5")
	_, err = Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for float max, got nil")
	}

	t.Setenv("TEST_RATIO", "0.5")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Ratio != 0.5 {
		t.Errorf("expected Ratio 0.5, got %v", cfg.Ratio)
	}
}

func TestValidationUintMinMax(t *testing.T) {
	type Config struct {
		Count uint `env:"TEST_UINT_COUNT" validate:"min=1,max=100"`
	}

	t.Setenv("TEST_UINT_COUNT", "0")
	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for uint min, got nil")
	}

	t.Setenv("TEST_UINT_COUNT", "50")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Count != 50 {
		t.Errorf("expected Count 50, got %d", cfg.Count)
	}
}

func TestErrorReportImplementsError(t *testing.T) {
	report := &ErrorReport{}
	report.AddError(FieldError{
		Path:    "Port",
		Kind:    ErrorKindValidation,
		Message: "must be positive",
	})
	if report.Error() == "" {
		t.Errorf("expected non-empty error string from ErrorReport.Error()")
	}
	if report.IsEmpty() {
		t.Errorf("expected IsEmpty() to return false when there are errors")
	}
	empty := &ErrorReport{}
	if !empty.IsEmpty() {
		t.Errorf("expected IsEmpty() to return true for empty report")
	}
}

func TestExplainVariousErrors(t *testing.T) {
	// Test Explain with a regular error
	err := fmt.Errorf("plain error")
	result := Explain(err)
	if !strings.Contains(result, "plain error") {
		t.Errorf("expected 'plain error' in explained result, got: %s", result)
	}

	// Test Explain with an ErrorReport
	report := &ErrorReport{}
	report.AddError(FieldError{
		Path:    "Host",
		Kind:    ErrorKindValidation,
		Rule:    "required",
		Message: "field is required",
		Source:  "env HOST",
	})
	result = Explain(report)
	if !strings.Contains(result, "Host") {
		t.Errorf("expected 'Host' in explained report, got: %s", result)
	}
}

func TestValidationPtrField(t *testing.T) {
	type Config struct {
		Count *int `env:"TEST_PTR_COUNT" validate:"required"`
	}
	// nil pointer should satisfy required check differently
	_, err := Load[Config](FromEnv())
	// Just check it doesn't panic
	_ = err
}

func TestLoadMissingYAMLFile(t *testing.T) {
	type Config struct {
		Port int `yaml:"port"`
	}

	_, err := Load[Config](FromYAML("testdata/nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing YAML file, got nil")
	}
}

func TestLoadMissingJSONFile(t *testing.T) {
	type Config struct {
		Port int `json:"port"`
	}

	_, err := Load[Config](FromJSON("testdata/nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing JSON file, got nil")
	}
}

func TestLoadMissingTOMLFile(t *testing.T) {
	type Config struct {
		Port int `toml:"port"`
	}

	_, err := Load[Config](FromTOML("testdata/nonexistent.toml"))
	if err == nil {
		t.Fatal("expected error for missing TOML file, got nil")
	}
}

func TestNestedStructSnakeCaseFallback(t *testing.T) {
	// When a nested struct parent has no yaml tag, lookup uses snake_case of field name
	// "Database" → "database" to match config.yaml's [database] key
	type DatabaseConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	type Config struct {
		Database DatabaseConfig // No yaml tag: falls back to "database" via toSnakeCase
	}

	cfg, err := Load[Config](FromYAML("testdata/config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host 'localhost', got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port 5432, got %d", cfg.Database.Port)
	}
}

func TestFieldValueToStringTypes(t *testing.T) {
	type Config struct {
		Active bool    `env:"TEST_BOOL_FMT" validate:"required"`
		Ratio  float32 `env:"TEST_FLOAT_FMT" validate:"required"`
		Tags   []string `env:"TEST_SLICE_FMT"`
	}

	t.Setenv("TEST_BOOL_FMT", "true")
	t.Setenv("TEST_FLOAT_FMT", "3.14")
	t.Setenv("TEST_SLICE_FMT", "a,b")

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Active {
		t.Errorf("expected Active true, got %v", cfg.Active)
	}
}

func TestValidationBoolRequired(t *testing.T) {
	type Config struct {
		// bool zero value is false, which is falsy for required check
		Flag bool `env:"TEST_BOOL_REQUIRED" validate:"required"`
	}

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for required bool (false), got nil")
	}
}

func TestValidationSliceRequired(t *testing.T) {
	type Config struct {
		Items []string `env:"TEST_SLICE_REQUIRED" validate:"required"`
	}

	_, err := Load[Config](FromEnv())
	if err == nil {
		t.Fatal("expected validation error for required empty slice, got nil")
	}
}

func TestParseUintTypes(t *testing.T) {
	p := NewParser()

	tests := []struct {
		input    string
		typ      reflect.Type
		expected any
		wantErr  bool
	}{
		{"100", reflect.TypeOf(uint(0)), uint(100), false},
		{"200", reflect.TypeOf(uint16(0)), uint16(200), false},
		{"300", reflect.TypeOf(uint32(0)), uint32(300), false},
		{"400", reflect.TypeOf(uint64(0)), uint64(400), false},
		{"invalid", reflect.TypeOf(uint(0)), nil, true},
	}

	for _, tc := range tests {
		result, err := p.Parse(tc.input, tc.typ)
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q, %v): got error %v, wantErr %v", tc.input, tc.typ, err, tc.wantErr)
			continue
		}
		if err == nil && result != tc.expected {
			t.Errorf("Parse(%q, %v): expected %v, got %v", tc.input, tc.typ, tc.expected, result)
		}
	}
}

func TestEnvPrefixSimple(t *testing.T) {
	type Server struct {
		Port int `env:"PORT"`
	}
	type Config struct {
		Server Server `prefix:"APP_"`
	}

	t.Setenv("APP_PORT", "9000")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected Server.Port=9000, got %d", cfg.Server.Port)
	}
}

func TestEnvPrefixHierarchical(t *testing.T) {
	type Database struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}
	type Config struct {
		Database Database `prefix:"DB_"`
	}

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host=localhost, got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port=5432, got %d", cfg.Database.Port)
	}
}

func TestEnvPrefixMultiLevel(t *testing.T) {
	type Cache struct {
		TTL int `env:"TTL"`
	}
	type Database struct {
		Host  string `env:"HOST"`
		Cache Cache  `prefix:"CACHE_"`
	}
	type Config struct {
		Database Database `prefix:"DB_"`
	}

	t.Setenv("DB_HOST", "postgres.local")
	t.Setenv("DB_CACHE_TTL", "3600")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "postgres.local" {
		t.Errorf("expected Database.Host=postgres.local, got %q", cfg.Database.Host)
	}
	if cfg.Database.Cache.TTL != 3600 {
		t.Errorf("expected Database.Cache.TTL=3600, got %d", cfg.Database.Cache.TTL)
	}
}

func TestEnvPrefixWithDefault(t *testing.T) {
	type Server struct {
		Timeout int `env:"TIMEOUT" default:"30"`
	}
	type Config struct {
		Server Server `prefix:"APP_"`
	}

	if _, ok := os.LookupEnv("APP_TIMEOUT"); ok {
		t.Skip("APP_TIMEOUT already set")
	}

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Timeout != 30 {
		t.Errorf("expected Server.Timeout=30 (default), got %d", cfg.Server.Timeout)
	}
}

func TestEnvPrefixOverride(t *testing.T) {
	type Logger struct {
		Level string `env:"LOG_LEVEL" default:"info"`
	}
	type Config struct {
		Logger Logger `prefix:"APP_"`
	}

	t.Setenv("APP_LOG_LEVEL", "debug")
	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("expected Logger.Level=debug, got %q", cfg.Logger.Level)
	}
}
