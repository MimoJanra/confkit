package confkit

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestValidation(t *testing.T) {
	t.Run("required/missing_string", func(t *testing.T) {
		type cfg struct {
			Host string `env:"TEST_REQUIRED_HOST" validate:"required"`
		}
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required field")
		}
		report := assertReport(t, err)
		if report.Errors[0].Kind != ErrorKindValidation || report.Errors[0].Path != "Host" {
			t.Errorf("unexpected error: %+v", report.Errors[0])
		}
	})

	t.Run("required/bool_false", func(t *testing.T) {
		type cfg struct {
			Flag bool `env:"TEST_BOOL_REQUIRED" validate:"required"`
		}
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required bool (false)")
		}
	})

	t.Run("required/empty_slice", func(t *testing.T) {
		type cfg struct {
			Items []string `env:"TEST_SLICE_REQUIRED" validate:"required"`
		}
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required empty slice")
		}
	})

	t.Run("required/bool_true_passes", func(t *testing.T) {
		type cfg struct {
			Flag bool `env:"VAL_BOOL_FLAG" validate:"required"`
		}
		_ = os.Setenv("VAL_BOOL_FLAG", "true")
		t.Cleanup(func() { _ = os.Unsetenv("VAL_BOOL_FLAG") })
		_, err := Load[cfg](FromEnv())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("min/int_below", func(t *testing.T) {
		type cfg struct {
			Port int `env:"TEST_MIN_PORT" validate:"min=1"`
		}
		t.Setenv("TEST_MIN_PORT", "0")
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for min constraint")
		}
		_ = assertReport(t, err)
	})

	t.Run("max/int_above", func(t *testing.T) {
		type cfg struct {
			Port int `env:"TEST_MAX_PORT" validate:"max=65535"`
		}
		t.Setenv("TEST_MAX_PORT", "99999")
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for max constraint")
		}
	})

	t.Run("multiple_rules_pass", func(t *testing.T) {
		type cfg struct {
			Port int `env:"TEST_MULTI_PORT" validate:"required,min=1,max=65535"`
		}
		t.Setenv("TEST_MULTI_PORT", "8080")
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Port != 8080 {
			t.Errorf("expected 8080, got %d", c.Port)
		}
	})

	t.Run("string/min_max", func(t *testing.T) {
		type cfg struct {
			Name string `env:"TEST_NAME_LEN" validate:"min=3,max=10"`
		}
		t.Setenv("TEST_NAME_LEN", "ab")
		if _, err := Load[cfg](FromEnv()); err == nil {
			t.Fatal("expected error for string too short")
		}
		t.Setenv("TEST_NAME_LEN", "this-is-too-long-for-max")
		if _, err := Load[cfg](FromEnv()); err == nil {
			t.Fatal("expected error for string too long")
		}
		t.Setenv("TEST_NAME_LEN", "hello")
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "hello" {
			t.Errorf("expected 'hello', got %q", c.Name)
		}
	})

	t.Run("float/min_max", func(t *testing.T) {
		type cfg struct {
			Ratio float64 `env:"TEST_RATIO" validate:"min=0,max=1"`
		}
		t.Setenv("TEST_RATIO", "-0.5")
		if _, err := Load[cfg](FromEnv()); err == nil {
			t.Fatal("expected error for float below min")
		}
		t.Setenv("TEST_RATIO", "1.5")
		if _, err := Load[cfg](FromEnv()); err == nil {
			t.Fatal("expected error for float above max")
		}
		t.Setenv("TEST_RATIO", "0.5")
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Ratio != 0.5 {
			t.Errorf("expected 0.5, got %v", c.Ratio)
		}
	})

	t.Run("uint/min_max", func(t *testing.T) {
		type cfg struct {
			Count uint `env:"TEST_UINT_COUNT" validate:"min=1,max=100"`
		}
		t.Setenv("TEST_UINT_COUNT", "0")
		if _, err := Load[cfg](FromEnv()); err == nil {
			t.Fatal("expected error for uint below min")
		}
		t.Setenv("TEST_UINT_COUNT", "50")
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Count != 50 {
			t.Errorf("expected 50, got %d", c.Count)
		}
	})

	t.Run("oneof/invalid", func(t *testing.T) {
		type cfg struct {
			Level string `env:"TEST_LOGLEVEL" validate:"oneof=debug,info,warn,error"`
		}
		t.Setenv("TEST_LOGLEVEL", "invalid")
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error for oneof")
		}
		report := assertReport(t, err)
		if report.Errors[0].Rule != "oneof" {
			t.Errorf("expected rule 'oneof', got %s", report.Errors[0].Rule)
		}
	})

	t.Run("oneof/valid", func(t *testing.T) {
		type cfg struct {
			Level string `env:"TEST_LOGLEVEL_VALID" validate:"oneof=debug,info,warn,error"`
		}
		t.Setenv("TEST_LOGLEVEL_VALID", "info")
		c, err := Load[cfg](FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Level != "info" {
			t.Errorf("expected 'info', got %q", c.Level)
		}
	})

	t.Run("oneof/with_required", func(t *testing.T) {
		type cfg struct {
			Level string `env:"TEST_LEVEL_REQUIRED" validate:"required,oneof=debug,info"`
		}
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected validation error")
		}
		report := assertReport(t, err)
		if report.Errors[0].Rule != "required" {
			t.Errorf("expected rule 'required', got %s", report.Errors[0].Rule)
		}
	})

	t.Run("secret_redaction", func(t *testing.T) {
		report := &ErrorReport{}
		report.AddError(FieldError{
			Path:    "APIKey",
			Kind:    ErrorKindValidation,
			Message: "invalid format",
			Value:   "super-secret-12345",
			Secret:  true,
		})
		formatted := report.Format()
		if !strings.Contains(formatted, "***REDACTED***") {
			t.Errorf("expected secret to be redacted: %s", formatted)
		}
		if strings.Contains(formatted, "super-secret-12345") {
			t.Errorf("secret value leaked: %s", formatted)
		}
	})

	t.Run("custom", func(t *testing.T) {
		type cfg struct {
			Count int `env:"TEST_COUNT" validate:"positive"`
		}
		t.Setenv("TEST_COUNT", "-5")
		_, err := LoadWithOptions[cfg](
			WithSource(FromEnv()),
			WithValidator("positive", func(v reflect.Value) error {
				if v.Kind() == reflect.Int && v.Int() <= 0 {
					return fmt.Errorf("must be positive")
				}
				return nil
			}),
		)
		if err == nil {
			t.Fatal("expected custom validator error")
		}
		report := assertReport(t, err)
		if report.Errors[0].Rule != "positive" {
			t.Errorf("expected rule 'positive', got %s", report.Errors[0].Rule)
		}
		if !strings.Contains(report.Errors[0].Message, "must be positive") {
			t.Errorf("unexpected message: %s", report.Errors[0].Message)
		}
	})

	t.Run("nested", func(t *testing.T) {
		type DB struct {
			Host string `yaml:"host" validate:"required"`
			Port int    `yaml:"port" validate:"min=1,max=65535"`
		}
		type cfg struct {
			Port int `yaml:"port" default:"8080"`
			DB   DB  `yaml:"database"`
		}
		c, err := Load[cfg](FromYAML("testdata/config.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DB.Port != 5432 {
			t.Errorf("expected DB.Port 5432, got %d", c.DB.Port)
		}
	})

	t.Run("parse_error_report", func(t *testing.T) {
		type cfg struct {
			Port int `env:"BAD_PORT"`
		}
		t.Setenv("BAD_PORT", "not-a-number")
		_, err := Load[cfg](FromEnv())
		if err == nil {
			t.Fatal("expected error")
		}
		report := assertReport(t, err)
		if report.Errors[0].Kind != ErrorKindParse || report.Errors[0].Path != "Port" {
			t.Errorf("unexpected error: %+v", report.Errors[0])
		}
	})
}

func TestValidationInternals(t *testing.T) {
	t.Run("port/uint", func(t *testing.T) {
		v := NewValidator()
		field := FieldInfo{Path: "Port", Name: "Port"}
		rule := ValidationRule{Name: "port"}

		fe := v.validateField(reflect.ValueOf(uint(8080)), field, rule)
		if fe.Message != "" {
			t.Errorf("expected no error for valid uint port, got: %s", fe.Message)
		}
		fe = v.validateField(reflect.ValueOf(uint(0)), field, rule)
		if fe.Message == "" {
			t.Error("expected error for port 0 (uint)")
		}
	})

	t.Run("port/string", func(t *testing.T) {
		v := NewValidator()
		field := FieldInfo{Path: "Port", Name: "Port"}
		rule := ValidationRule{Name: "port"}

		fe := v.validateField(reflect.ValueOf("443"), field, rule)
		if fe.Message != "" {
			t.Errorf("expected no error for valid string port, got: %s", fe.Message)
		}
		fe = v.validateField(reflect.ValueOf("notanumber"), field, rule)
		if fe.Message == "" {
			t.Error("expected error for non-numeric string port")
		}
	})

	t.Run("str_check_non_string", func(t *testing.T) {
		v := NewValidator()
		field := FieldInfo{Path: "X", Name: "X"}
		fe := v.validateField(reflect.ValueOf(42), field, ValidationRule{Name: "email"})
		if fe.Message != "" {
			t.Errorf("expected no error for non-string field with string validator, got: %s", fe.Message)
		}
	})

	t.Run("parse_rules_oneof", func(t *testing.T) {
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
	})
}

func assertReport(t *testing.T, err error) *ErrorReport {
	t.Helper()
	report, ok := err.(*ErrorReport)
	if !ok {
		t.Fatalf("expected *ErrorReport, got %T: %v", err, err)
	}
	if len(report.Errors) == 0 {
		t.Fatal("expected at least one error in report")
	}
	return report
}
