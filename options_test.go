package confkit

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadWithOptionsBasic(t *testing.T) {
	t.Setenv("PORT", "3000")

	type Config struct {
		Port int `env:"PORT"`
	}

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("Expected Port=3000, got %d", cfg.Port)
	}
}

func TestLoadWithMultipleSources(t *testing.T) {
	yamlContent := `
Port: 8080
Host: localhost
`
	tmpFile := writeTempYAML(t, yamlContent)
	defer func() { _ = os.Remove(tmpFile) }()

	t.Setenv("HOST", "0.0.0.0")

	type Config struct {
		Port int    `yaml:"Port" env:"PORT"`
		Host string `yaml:"Host" env:"HOST"`
	}

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithSource(FromYAML(tmpFile)),
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

	cfg, err := LoadWithOptions[Config](
		WithInterpolationMaxDepth(5),
	)
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
		portVal := v.Int()
		if portVal < 1 || portVal > 65535 {
			return ErrorPortOutOfRange{portVal}
		}
		return nil
	}

	t.Setenv("PORT", "99999")

	_, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithValidator("custom_range", customValidator),
	)
	if err == nil {
		t.Fatal("Expected validation error for port out of range")
	}
}

type ErrorPortOutOfRange struct {
	port int64
}

func (e ErrorPortOutOfRange) Error() string {
	return "port out of range"
}

func writeTempYAML(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return f.Name()
}

func TestLoadWithMiddleware(t *testing.T) {
	t.Setenv("NAME", "  john  ")

	type Config struct {
		Name string `env:"NAME"`
	}

	trimMiddleware := func(field FieldInfo, value string) (string, error) {
		return strings.TrimSpace(value), nil
	}

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithMiddleware(trimMiddleware),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Name != "john" {
		t.Errorf("Expected Name='john', got '%s'", cfg.Name)
	}
}

func writeTempJSON(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return f.Name()
}

func writeTempTOML(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "test-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return f.Name()
}

func TestWithLoadHook(t *testing.T) {
	type Config struct {
		Port int `env:"HOOK_PORT" default:"8080"`
	}

	hookCalled := false
	hookSuccess := false
	var hookDuration time.Duration

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithLoadHook(func(success bool, duration time.Duration, errCount int) {
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

	hookCalled := false
	hookSuccess := false
	hookErrCount := 0

	t.Setenv("HOOK_ERROR_PORT", "invalid")

	_, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithLoadHook(func(success bool, duration time.Duration, errCount int) {
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

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithModelValidator[Config](func(c *Config) error {
			if c.Username == "" {
				return ErrorInvalidUser{"empty"}
			}
			if len(c.Password) < 8 {
				return ErrorWeakPassword{len(c.Password)}
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

	_, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithModelValidator[Config](func(c *Config) error {
			if len(c.Password) < 8 {
				return ErrorWeakPassword{len(c.Password)}
			}
			return nil
		}),
	)

	if err == nil {
		t.Fatal("Expected validation error for weak password")
	}
}

type ErrorInvalidUser struct {
	value string
}

func (e ErrorInvalidUser) Error() string {
	return "invalid username: " + e.value
}

type ErrorWeakPassword struct {
	length int
}

func (e ErrorWeakPassword) Error() string {
	return "password too weak: " + string(rune(e.length)) + " chars"
}
