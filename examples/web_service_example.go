package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MimoJanra/confkit"
)

// WebServiceConfig demonstrates a typical web service configuration
// with nested structs, validation, and multiple sources.
type WebServiceConfig struct {
	App struct {
		Name    string `env:"APP_NAME" default:"myapp" validate:"required"`
		Version string `env:"APP_VERSION" default:"0.1.0"`
		Env     string `env:"ENV" default:"development" validate:"oneof=development,staging,production"`
	} `prefix:"APP_"`

	Server struct {
		Host     string        `env:"HOST" default:"0.0.0.0" validate:"required"`
		Port     int           `env:"PORT" default:"8080" validate:"min=1,max=65535"`
		ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"15s" validate:"min=1s,max=5m"`
		WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"15s" validate:"min=1s,max=5m"`
	} `prefix:"SERVER_"`

	Database struct {
		Host     string        `env:"HOST" validate:"required"`
		Port     int           `env:"PORT" default:"5432" validate:"min=1,max=65535"`
		User     string        `env:"USER" validate:"required"`
		Password string        `env:"PASSWORD" validate:"required" secret:"true"`
		Name     string        `env:"NAME" validate:"required"`
		MaxConn  int           `env:"MAX_CONN" default:"20" validate:"min=1,max=100"`
		Timeout  time.Duration `env:"TIMEOUT" default:"30s" validate:"min=5s,max=5m"`
	} `prefix:"DB_"`

	Cache struct {
		Enabled bool          `env:"ENABLED" default:"true"`
		TTL     time.Duration `env:"TTL" default:"1h" validate:"min=1m,max=24h"`
		MaxSize int           `env:"MAX_SIZE" default:"1000" validate:"min=1"`
	} `prefix:"CACHE_"`

	Logging struct {
		Level  string `env:"LEVEL" default:"info" validate:"oneof=debug,info,warn,error"`
		Format string `env:"FORMAT" default:"json" validate:"oneof=json,text"`
	} `prefix:"LOG_"`

	Features struct {
		EnableMetrics   bool `env:"ENABLE_METRICS" default:"true"`
		EnableProfiling bool `env:"ENABLE_PROFILING" default:"false"`
		EnableTracing   bool `env:"ENABLE_TRACING" default:"true"`
	} `prefix:"FEATURE_"`
}

// ExampleWebService shows how to load and use web service configuration.
func ExampleWebService() error {
	// Load config from environment variables and config file
	cfg, err := confkit.Load[WebServiceConfig](
		confkit.FromEnv(),
		confkit.FromYAMLOptional("config.yaml"),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use the configuration
	log.Printf("Starting %s v%s in %s mode\n", cfg.App.Name, cfg.App.Version, cfg.App.Env)
	log.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Database: %s@%s:%d/%s\n", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	log.Printf("Cache enabled: %v (TTL: %v)\n", cfg.Cache.Enabled, cfg.Cache.TTL)
	log.Printf("Logging level: %s (format: %s)\n", cfg.Logging.Level, cfg.Logging.Format)

	return nil
}
