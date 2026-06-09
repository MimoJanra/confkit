package confkit_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	confkit "github.com/MimoJanra/confkit"
)

type errPortOutOfRange struct{ port int64 }

func (e errPortOutOfRange) Error() string { return fmt.Sprintf("port %d out of range", e.port) }

type errInvalidUser struct{ value string }

func (e errInvalidUser) Error() string { return "invalid username: " + e.value }

type errWeakPassword struct{ length int }

func (e errWeakPassword) Error() string {
	return fmt.Sprintf("password too weak: %d chars", e.length)
}

func TestLoadWithOptionsBasic(t *testing.T) {
	t.Setenv("PORT", "3000")

	type Config struct {
		Port int `env:"PORT"`
	}
	cfg, err := confkit.LoadWithOptions[Config](confkit.WithSource(confkit.FromEnv()))
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("Expected Port=3000, got %d", cfg.Port)
	}
}

func TestLoadWithMultipleSources(t *testing.T) {
	yamlContent := "Port: 8080\nHost: localhost\n"
	tmpFile := writeTempYAML(t, yamlContent)
	defer func() { _ = os.Remove(tmpFile) }()

	t.Setenv("HOST", "0.0.0.0")

	type Config struct {
		Port int    `yaml:"Port" env:"PORT"`
		Host string `yaml:"Host" env:"HOST"`
	}
	cfg, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithSource(confkit.FromYAML(tmpFile)),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected Host='0.0.0.0', got '%s'", cfg.Host)
	}
}

func TestLoadWithInterpolationMaxDepth(t *testing.T) {
	type Config struct {
		Value string `default:"${A|fallback}"`
	}
	cfg, err := confkit.LoadWithOptions[Config](confkit.WithInterpolationMaxDepth(5))
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Value != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", cfg.Value)
	}
}

func TestLoadWithValidator(t *testing.T) {
	type Config struct {
		Port int `env:"PORT" validate:"custom_range"`
	}
	customValidator := func(v reflect.Value) error {
		if v.Kind() != reflect.Int {
			return nil
		}
		if portVal := v.Int(); portVal < 1 || portVal > 65535 {
			return errPortOutOfRange{portVal}
		}
		return nil
	}
	t.Setenv("PORT", "99999")
	_, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithValidator("custom_range", customValidator),
	)
	if err == nil {
		t.Fatal("Expected validation error for port out of range")
	}
}

func TestLoadWithMiddleware(t *testing.T) {
	t.Setenv("NAME", "  john  ")

	type Config struct {
		Name string `env:"NAME"`
	}
	trimMiddleware := func(field confkit.FieldInfo, value string) (string, error) {
		return strings.TrimSpace(value), nil
	}
	cfg, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithMiddleware(trimMiddleware),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Name != "john" {
		t.Errorf("Expected Name='john', got '%s'", cfg.Name)
	}
}

func TestWithLoadHook(t *testing.T) {
	type Config struct {
		Port int `env:"HOOK_PORT" default:"8080"`
	}
	hookCalled, hookSuccess := false, false
	var hookDuration time.Duration

	cfg, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithLoadHook(func(success bool, duration time.Duration, errCount int) {
			hookCalled = true
			hookSuccess = success
			hookDuration = duration
		}),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if !hookCalled {
		t.Error("Expected load hook to be called")
	}
	if !hookSuccess {
		t.Error("Expected hook success=true")
	}
	if hookDuration < 0 {
		t.Errorf("Expected non-negative duration, got %v", hookDuration)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", cfg.Port)
	}
}

func TestWithLoadHookOnError(t *testing.T) {
	type Config struct {
		Port int `env:"HOOK_ERROR_PORT" validate:"min=1"`
	}
	hookCalled, hookSuccess := false, false
	hookErrCount := 0
	t.Setenv("HOOK_ERROR_PORT", "invalid")

	_, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithLoadHook(func(success bool, duration time.Duration, errCount int) {
			hookCalled = true
			hookSuccess = success
			hookErrCount = errCount
		}),
	)
	if err == nil {
		t.Fatal("Expected error from invalid port")
	}
	if !hookCalled {
		t.Error("Expected load hook to be called even on error")
	}
	if hookSuccess {
		t.Error("Expected hook success=false when error occurs")
	}
	if hookErrCount <= 0 {
		t.Errorf("Expected positive error count, got %d", hookErrCount)
	}
}

func TestWithModelValidatorSuccess(t *testing.T) {
	type Config struct {
		Username string `env:"MODEL_USER"`
		Password string `env:"MODEL_PASS"`
	}
	t.Setenv("MODEL_USER", "alice")
	t.Setenv("MODEL_PASS", "secret123")

	cfg, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithModelValidator(func(c *Config) error {
			if c.Username == "" {
				return errInvalidUser{"empty"}
			}
			if len(c.Password) < 8 {
				return errWeakPassword{len(c.Password)}
			}
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Username != "alice" {
		t.Errorf("Expected Username='alice', got %q", cfg.Username)
	}
}

func TestWithModelValidatorFails(t *testing.T) {
	type Config struct {
		Username string `env:"MODEL_USER2"`
		Password string `env:"MODEL_PASS2"`
	}
	t.Setenv("MODEL_USER2", "bob")
	t.Setenv("MODEL_PASS2", "weak")

	_, err := confkit.LoadWithOptions[Config](
		confkit.WithSource(confkit.FromEnv()),
		confkit.WithModelValidator(func(c *Config) error {
			if len(c.Password) < 8 {
				return errWeakPassword{len(c.Password)}
			}
			return nil
		}),
	)
	if err == nil {
		t.Fatal("Expected validation error for weak password")
	}
}
