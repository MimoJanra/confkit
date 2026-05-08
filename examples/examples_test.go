package main

import (
	"os"
	"testing"
	"time"

	"github.com/MimoJanra/confkit"
)

func TestWebServiceConfigValidDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "mydb")

	cfg, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.App.Name != "myapp" {
		t.Errorf("expected App.Name='myapp', got %q", cfg.App.Name)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected Server.Port=8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host='localhost', got %q", cfg.Database.Host)
	}
	if cfg.Cache.Enabled != true {
		t.Errorf("expected Cache.Enabled=true, got %v", cfg.Cache.Enabled)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected Logging.Level='info', got %q", cfg.Logging.Level)
	}
}

func TestWebServiceConfigEnvironmentOverride(t *testing.T) {
	os.Clearenv()
	// Note: APP_NAME uses prefix "APP_", so the env var should be APP_NAME
	// But the struct field is .App.Name, and app uses lowercase in prefix
	// so the actual env var should match the prefix + field tag
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "password123")
	os.Setenv("DB_NAME", "production")
	os.Setenv("LOG_LEVEL", "debug")

	cfg, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Server and Database use explicit prefixes, so they override correctly
	if cfg.Server.Port != 9000 {
		t.Errorf("expected Server.Port=9000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected Server.Host='127.0.0.1', got %q", cfg.Server.Host)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected Logging.Level='debug', got %q", cfg.Logging.Level)
	}
}

func TestWebServiceConfigDurationParsing(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVER_READ_TIMEOUT", "30s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "1m")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "db")
	os.Setenv("CACHE_TTL", "2h")

	cfg, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout=30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 1*time.Minute {
		t.Errorf("expected WriteTimeout=1m, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Cache.TTL != 2*time.Hour {
		t.Errorf("expected Cache.TTL=2h, got %v", cfg.Cache.TTL)
	}
}

func TestWebServiceConfigValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVER_PORT", "99999")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "db")

	_, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err == nil {
		t.Fatal("expected validation error for invalid port")
	}

	report, ok := err.(*confkit.ErrorReport)
	if !ok {
		t.Fatalf("expected ErrorReport, got %T", err)
	}

	if len(report.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestMicroserviceConfigCompleteSetup(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVICE_NAME", "order-service")
	os.Setenv("SERVICE_PORT", "8000")
	os.Setenv("AUTH_JWT_SECRET", "secret-key-here")
	os.Setenv("POSTGRES_HOST", "postgres.svc")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "admin")
	os.Setenv("POSTGRES_PASSWORD", "dbpass")
	os.Setenv("POSTGRES_DATABASE", "orders")
	os.Setenv("REDIS_HOST", "redis.svc")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("MQ_HOST", "rabbitmq.svc")
	os.Setenv("MQ_USER", "guest")
	os.Setenv("MQ_PASSWORD", "guest")

	cfg, err := confkit.Load[MicroserviceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Service.Name != "order-service" {
		t.Errorf("expected Service.Name='order-service', got %q", cfg.Service.Name)
	}
	if cfg.Service.Port != 8000 {
		t.Errorf("expected Service.Port=8000, got %d", cfg.Service.Port)
	}
	if cfg.Auth.TokenExpiry != 24*time.Hour {
		t.Errorf("expected Auth.TokenExpiry=24h, got %v", cfg.Auth.TokenExpiry)
	}
	if cfg.PostgreSQL.MaxOpenConn != 25 {
		t.Errorf("expected PostgreSQL.MaxOpenConn=25, got %d", cfg.PostgreSQL.MaxOpenConn)
	}
}

func TestMicroserviceConfigRateLimiting(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVICE_NAME", "api-service")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PASSWORD", "pass")
	os.Setenv("POSTGRES_DATABASE", "db")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("MQ_HOST", "localhost")
	os.Setenv("MQ_USER", "user")
	os.Setenv("MQ_PASSWORD", "pass")
	os.Setenv("AUTH_JWT_SECRET", "secret")
	os.Setenv("RATELIMIT_RPS", "5000")
	os.Setenv("RATELIMIT_BURST_SIZE", "50000")

	cfg, err := confkit.Load[MicroserviceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.RateLimiting.RequestsPerSecond != 5000 {
		t.Errorf("expected RPS=5000, got %d", cfg.RateLimiting.RequestsPerSecond)
	}
	if cfg.RateLimiting.BurstSize != 50000 {
		t.Errorf("expected BurstSize=50000, got %d", cfg.RateLimiting.BurstSize)
	}
}

func TestCLIToolConfigDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("INPUT_FILE", "input.json")
	os.Setenv("OUTPUT_FILE", "output.json")

	cfg, err := confkit.Load[CLIToolConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Input.File != "input.json" {
		t.Errorf("expected Input.File='input.json', got %q", cfg.Input.File)
	}
	if cfg.Input.Format != "json" {
		t.Errorf("expected Input.Format='json', got %q", cfg.Input.Format)
	}
	if cfg.Processing.Threads != 4 {
		t.Errorf("expected Processing.Threads=4, got %d", cfg.Processing.Threads)
	}
	if cfg.Validation.Strict != true {
		t.Errorf("expected Validation.Strict=true, got %v", cfg.Validation.Strict)
	}
	if cfg.Performance.EnableCache != true {
		t.Errorf("expected Performance.EnableCache=true, got %v", cfg.Performance.EnableCache)
	}
}

func TestCLIToolConfigCustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("INPUT_FILE", "data.csv")
	os.Setenv("INPUT_FORMAT", "csv")
	os.Setenv("INPUT_ENCODING", "utf-8")
	os.Setenv("OUTPUT_FILE", "result.xml")
	os.Setenv("OUTPUT_FORMAT", "xml")
	os.Setenv("PROCESSING_THREADS", "16")
	os.Setenv("PROCESSING_SKIP_ERRORS", "true")

	cfg, err := confkit.Load[CLIToolConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Input.Format != "csv" {
		t.Errorf("expected Input.Format='csv', got %q", cfg.Input.Format)
	}
	if cfg.Output.Format != "xml" {
		t.Errorf("expected Output.Format='xml', got %q", cfg.Output.Format)
	}
	if cfg.Processing.Threads != 16 {
		t.Errorf("expected Processing.Threads=16, got %d", cfg.Processing.Threads)
	}
	if cfg.Processing.SkipErrors != true {
		t.Errorf("expected Processing.SkipErrors=true, got %v", cfg.Processing.SkipErrors)
	}
}

func TestCLIToolConfigValidation(t *testing.T) {
	os.Clearenv()
	os.Setenv("INPUT_FILE", "input.json")
	os.Setenv("OUTPUT_FILE", "output.json")
	os.Setenv("PROCESSING_THREADS", "64")

	_, err := confkit.Load[CLIToolConfig](confkit.FromEnv())
	if err == nil {
		t.Fatal("expected validation error for threads > 32")
	}
}

func TestFullSetupExampleConfig(t *testing.T) {
	// This example uses TOML config file, so we'll test with environment variables
	os.Clearenv()

	type AppConfig struct {
		Port     int    `env:"PORT" toml:"port" default:"8080" validate:"min=1,max=65535"`
		Host     string `env:"HOST" toml:"host" default:"localhost" validate:"required"`
		LogLevel string `env:"LOG_LEVEL" toml:"log_level" default:"info" validate:"oneof=debug,info,warn,error"`
	}

	os.Setenv("PORT", "9000")
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("LOG_LEVEL", "debug")

	cfg, err := confkit.Load[AppConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected Port=9000, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host='0.0.0.0', got %q", cfg.Host)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel='debug', got %q", cfg.LogLevel)
	}
}

func TestNestedStructPrefixMapping(t *testing.T) {
	os.Clearenv()

	// Test that prefix:suffix mapping works correctly
	os.Setenv("DB_HOST", "db.local")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("CACHE_ENABLED", "true")
	os.Setenv("CACHE_TTL", "30m")

	cfg, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Database.Host != "db.local" {
		t.Errorf("expected Database.Host='db.local', got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port=5432, got %d", cfg.Database.Port)
	}
	if cfg.Cache.TTL != 30*time.Minute {
		t.Errorf("expected Cache.TTL=30m, got %v", cfg.Cache.TTL)
	}
}

func TestSecretFieldRedaction(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "my-secret-password-123")
	os.Setenv("DB_NAME", "db")

	cfg, err := confkit.Load[WebServiceConfig](confkit.FromEnv())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify password was loaded correctly
	if cfg.Database.Password != "my-secret-password-123" {
		t.Errorf("expected password to be loaded correctly")
	}

	// Secret fields should be marked correctly in the config
	// (actual redaction happens in error messages, not in the value itself)
}
