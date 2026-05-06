package confkit

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestE2EEnvPrefixAndInterpolation(t *testing.T) {
	t.Setenv("APP_HOST", "localhost")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_DB_HOST", "db.local")
	t.Setenv("APP_DB_PORT", "5432")

	type DatabaseConfig struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}

	type Config struct {
		Host      string         `env:"HOST" prefix:"APP_"`
		Port      int            `env:"PORT" prefix:"APP_"`
		BaseURL   string         `default:"http://${Host}:${Port}"`
		Database  DatabaseConfig `prefix:"APP_DB_"`
		DbConnStr string         `default:"postgres://${Database.Host}:${Database.Port}"`
	}

	cfg, err := Load[Config](FromEnv())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host='localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", cfg.Port)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL='http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.Database.Host != "db.local" {
		t.Errorf("Expected Database.Host='db.local', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Expected Database.Port=5432, got %d", cfg.Database.Port)
	}
	if cfg.DbConnStr != "postgres://db.local:5432" {
		t.Errorf("Expected DbConnStr='postgres://db.local:5432', got '%s'", cfg.DbConnStr)
	}
}

func TestE2EFunctionalOptionsWithCustomValidator(t *testing.T) {
	t.Setenv("PORT", "8080")

	type Config struct {
		Port int `env:"PORT" validate:"port_range"`
	}

	portRangeValidator := func(v reflect.Value) error {
		if v.Kind() != reflect.Int {
			return nil
		}
		port := v.Int()
		if port < 1 || port > 65535 {
			return ErrorInvalidPort{port}
		}
		return nil
	}

	cfg, err := LoadWithOptions[Config](
		WithSource(FromEnv()),
		WithValidator("port_range", portRangeValidator),
	)
	if err != nil {
		t.Fatalf("LoadWithOptions failed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", cfg.Port)
	}
}

type ErrorInvalidPort struct {
	port int64
}

func (e ErrorInvalidPort) Error() string {
	return "port out of valid range"
}

func TestE2EMiddlewareAndInterpolation(t *testing.T) {
	t.Setenv("API_KEY", "  secret123  ")
	t.Setenv("API_URL", "https://api.example.com")

	type Config struct {
		APIKey string `env:"API_KEY" secret:"true"`
		APIURL string `env:"API_URL"`
		Token  string `default:"Bearer ${APIKey}"`
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

	if cfg.APIKey != "secret123" {
		t.Errorf("Expected APIKey='secret123', got '%s'", cfg.APIKey)
	}
	if cfg.Token != "Bearer secret123" {
		t.Errorf("Expected Token='Bearer secret123', got '%s'", cfg.Token)
	}
}

func TestE2EMultiSourceWithPrecedence(t *testing.T) {
	yamlContent := `
Port: 3000
Host: yaml.host
Database: yaml_db
`
	tmpFile := writeTempYAML(t, yamlContent)
	defer func() { _ = os.Remove(tmpFile) }()

	t.Setenv("PORT", "8000")
	t.Setenv("HOST", "env.host")

	type Config struct {
		Port     int    `yaml:"Port" env:"PORT"`
		Host     string `yaml:"Host" env:"HOST"`
		Database string `yaml:"Database"`
	}

	cfg, err := Load[Config](
		FromEnv(),
		FromYAML(tmpFile),
	)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 8000 {
		t.Errorf("Expected Port=8000 (from env, first-wins), got %d", cfg.Port)
	}
	if cfg.Host != "env.host" {
		t.Errorf("Expected Host='env.host' (from env, first-wins), got '%s'", cfg.Host)
	}
	if cfg.Database != "yaml_db" {
		t.Errorf("Expected Database='yaml_db' (from yaml, no env source), got '%s'", cfg.Database)
	}
}
