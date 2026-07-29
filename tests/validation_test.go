package confkit_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	confkit "github.com/MimoJanra/confkit"
)

func TestValidation(t *testing.T) {
	t.Run("required/missing_string", func(t *testing.T) {
		type cfg struct {
			Host string `env:"TEST_REQUIRED_HOST" validate:"required"`
		}
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required field")
		}
		report := assertReport(t, err)
		if report.Errors[0].Kind != confkit.ErrorKindValidation || report.Errors[0].Path != "Host" {
			t.Errorf("unexpected error: %+v", report.Errors[0])
		}
	})

	t.Run("required/bool_false", func(t *testing.T) {
		type cfg struct {
			Flag bool `env:"TEST_BOOL_REQUIRED" validate:"required"`
		}
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required bool (false)")
		}
	})

	t.Run("required/bool_true_passes", func(t *testing.T) {
		type cfg struct {
			Flag bool `env:"VAL_BOOL_FLAG" validate:"required"`
		}
		_ = os.Setenv("VAL_BOOL_FLAG", "true")
		t.Cleanup(func() { _ = os.Unsetenv("VAL_BOOL_FLAG") })
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("min/int_below", func(t *testing.T) {
		type cfg struct {
			Port int `env:"TEST_MIN_PORT" validate:"min=1"`
		}
		t.Setenv("TEST_MIN_PORT", "0")
		_, err := confkit.Load[cfg](confkit.FromEnv())
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
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for max constraint")
		}
	})

	t.Run("multiple_rules_pass", func(t *testing.T) {
		type cfg struct {
			Port int `env:"TEST_MULTI_PORT" validate:"required,min=1,max=65535"`
		}
		t.Setenv("TEST_MULTI_PORT", "8080")
		c, err := confkit.Load[cfg](confkit.FromEnv())
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
		if _, err := confkit.Load[cfg](confkit.FromEnv()); err == nil {
			t.Fatal("expected error for string too short")
		}
		t.Setenv("TEST_NAME_LEN", "this-is-too-long-for-max")
		if _, err := confkit.Load[cfg](confkit.FromEnv()); err == nil {
			t.Fatal("expected error for string too long")
		}
		t.Setenv("TEST_NAME_LEN", "hello")
		c, err := confkit.Load[cfg](confkit.FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "hello" {
			t.Errorf("expected 'hello', got %q", c.Name)
		}
	})

	t.Run("oneof/invalid", func(t *testing.T) {
		type cfg struct {
			Level string `env:"TEST_LOGLEVEL" validate:"oneof=debug,info,warn,error"`
		}
		t.Setenv("TEST_LOGLEVEL", "invalid")
		_, err := confkit.Load[cfg](confkit.FromEnv())
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
		c, err := confkit.Load[cfg](confkit.FromEnv())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Level != "info" {
			t.Errorf("expected 'info', got %q", c.Level)
		}
	})

	t.Run("secret_redaction", func(t *testing.T) {
		report := &confkit.ErrorReport{}
		report.AddError(confkit.FieldError{
			Path:    "APIKey",
			Kind:    confkit.ErrorKindValidation,
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
		_, err := confkit.LoadWithOptions[cfg](
			confkit.WithSource(confkit.FromEnv()),
			confkit.WithValidator("positive", func(v reflect.Value) error {
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
		c, err := confkit.Load[cfg](confkit.FromYAML("../testdata/config.yaml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DB.Port != 5432 {
			t.Errorf("expected DB.Port 5432, got %d", c.DB.Port)
		}
	})

	t.Run("required/uint_zero", func(t *testing.T) {
		type cfg struct {
			Count uint `env:"TEST_UINT_REQ" validate:"required"`
		}
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required uint (zero value)")
		}
	})

	t.Run("required/uint_nonzero_passes", func(t *testing.T) {
		type cfg struct {
			Count uint `env:"TEST_UINT_NZ" validate:"required"`
		}
		t.Setenv("TEST_UINT_NZ", "5")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("required/float_zero", func(t *testing.T) {
		type cfg struct {
			Rate float64 `env:"TEST_FLOAT_REQ" validate:"required"`
		}
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for required float64 (zero value)")
		}
	})

	t.Run("min/uint", func(t *testing.T) {
		type cfg struct {
			Count uint `env:"TEST_UINT_MIN" validate:"min=5"`
		}
		t.Setenv("TEST_UINT_MIN", "2")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for uint below min")
		}
	})

	t.Run("max/uint", func(t *testing.T) {
		type cfg struct {
			Count uint `env:"TEST_UINT_MAX" validate:"max=10"`
		}
		t.Setenv("TEST_UINT_MAX", "20")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for uint above max")
		}
	})

	t.Run("min/float", func(t *testing.T) {
		type cfg struct {
			Rate float64 `env:"TEST_FLOAT_MIN" validate:"min=1"`
		}
		t.Setenv("TEST_FLOAT_MIN", "0.5")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for float below min")
		}
	})

	t.Run("max/float", func(t *testing.T) {
		type cfg struct {
			Rate float64 `env:"TEST_FLOAT_MAX" validate:"max=10"`
		}
		t.Setenv("TEST_FLOAT_MAX", "15.5")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected validation error for float above max")
		}
	})

	t.Run("parse_error_report", func(t *testing.T) {
		type cfg struct {
			Port int `env:"BAD_PORT"`
		}
		t.Setenv("BAD_PORT", "not-a-number")
		_, err := confkit.Load[cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected error")
		}
		report := assertReport(t, err)
		if report.Errors[0].Kind != confkit.ErrorKindParse || report.Errors[0].Path != "Port" {
			t.Errorf("unexpected error: %+v", report.Errors[0])
		}
	})
}

func TestSecretValueNotLeakedInValidationMessage(t *testing.T) {
	t.Run("port", func(t *testing.T) {
		type Cfg struct {
			Token int `env:"TEST_SECRET_PORT" secret:"true" validate:"port"`
		}
		t.Setenv("TEST_SECRET_PORT", "99999")

		_, err := confkit.Load[Cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected a port validation error")
		}
		if out := confkit.Explain(err); strings.Contains(out, "99999") {
			t.Fatalf("secret value leaked into error message:\n%s", out)
		}
	})

	t.Run("len", func(t *testing.T) {
		type Cfg struct {
			Token string `env:"TEST_SECRET_LEN" secret:"true" validate:"len=5"`
		}
		t.Setenv("TEST_SECRET_LEN", "abcdefghij")

		_, err := confkit.Load[Cfg](confkit.FromEnv())
		if err == nil {
			t.Fatal("expected a len validation error")
		}
		if out := confkit.Explain(err); strings.Contains(out, "got 10") {
			t.Fatalf("secret length leaked into error message:\n%s", out)
		}
	})
}
