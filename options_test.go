package confkit

import (
	"os"
	"reflect"
	"strings"
	"testing"
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
		WithSource(FromYAML(tmpFile)),
		WithSource(FromEnv()),
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
