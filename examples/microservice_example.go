package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MimoJanra/confkit"
)

// MicroserviceConfig demonstrates a microservice with multiple
// dependencies and comprehensive validation.
type MicroserviceConfig struct {
	Service struct {
		Name     string `env:"NAME" validate:"required"`
		Port     int    `env:"PORT" default:"8000" validate:"min=1024,max=65535"`
		DebugKey string `env:"DEBUG_KEY" secret:"true"`
	} `prefix:"SERVICE_"`

	Auth struct {
		Enabled       bool          `env:"ENABLED" default:"true"`
		JWTSecret     string        `env:"JWT_SECRET" secret:"true" validate:"required"`
		TokenExpiry   time.Duration `env:"TOKEN_EXPIRY" default:"24h" validate:"min=1h,max=720h"`
		RefreshExpiry time.Duration `env:"REFRESH_EXPIRY" default:"168h" validate:"min=24h,max=720h"` // 7 days = 168 hours
	} `prefix:"AUTH_"`

	PostgreSQL struct {
		Host         string        `env:"HOST" validate:"required"`
		Port         int           `env:"PORT" default:"5432" validate:"min=1,max=65535"`
		User         string        `env:"USER" validate:"required"`
		Password     string        `env:"PASSWORD" validate:"required" secret:"true"`
		Database     string        `env:"DATABASE" validate:"required"`
		MaxOpenConn  int           `env:"MAX_OPEN_CONN" default:"25" validate:"min=1,max=100"`
		MaxIdleConn  int           `env:"MAX_IDLE_CONN" default:"5" validate:"min=0,max=50"`
		IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" default:"10m" validate:"min=1m,max=1h"`
		QueryTimeout time.Duration `env:"QUERY_TIMEOUT" default:"30s" validate:"min=1s,max=5m"`
	} `prefix:"POSTGRES_"`

	Redis struct {
		Host         string        `env:"HOST" validate:"required"`
		Port         int           `env:"PORT" default:"6379" validate:"min=1,max=65535"`
		Password     string        `env:"PASSWORD" secret:"true"`
		Database     int           `env:"DATABASE" default:"0" validate:"min=0,max=15"`
		MaxRetries   int           `env:"MAX_RETRIES" default:"3" validate:"min=0,max=10"`
		PoolSize     int           `env:"POOL_SIZE" default:"10" validate:"min=1,max=100"`
		ConnTimeout  time.Duration `env:"CONN_TIMEOUT" default:"5s" validate:"min=1s,max=30s"`
		ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"5s" validate:"min=1s,max=30s"`
		WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"5s" validate:"min=1s,max=30s"`
	} `prefix:"REDIS_"`

	MessageQueue struct {
		Type              string        `env:"TYPE" default:"rabbitmq" validate:"oneof=rabbitmq,kafka,nats"`
		Host              string        `env:"HOST" validate:"required"`
		Port              int           `env:"PORT" default:"5672" validate:"min=1,max=65535"`
		User              string        `env:"USER" validate:"required"`
		Password          string        `env:"PASSWORD" validate:"required" secret:"true"`
		ChannelBufferSize int           `env:"CHANNEL_BUFFER_SIZE" default:"100" validate:"min=1,max=10000"`
		PrefetchCount     int           `env:"PREFETCH_COUNT" default:"10" validate:"min=1,max=1000"`
		ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" default:"10s" validate:"min=1s,max=1m"`
	} `prefix:"MQ_"`

	Observability struct {
		LogLevel       string  `env:"LOG_LEVEL" default:"info" validate:"oneof=debug,info,warn,error"`
		LogFormat      string  `env:"LOG_FORMAT" default:"json" validate:"oneof=json,text"`
		MetricsPort    int     `env:"METRICS_PORT" default:"8001" validate:"min=1024,max=65535"`
		MetricsPath    string  `env:"METRICS_PATH" default:"/metrics"`
		TracingEnabled bool    `env:"TRACING_ENABLED" default:"true"`
		TracesSampled  float64 `env:"TRACES_SAMPLED" default:"0.1" validate:"min=0,max=1"`
	} `prefix:"OBS_"`

	FeatureFlags struct {
		EnableNewAPI      bool `env:"ENABLE_NEW_API" default:"false"`
		EnableBatchOps    bool `env:"ENABLE_BATCH_OPS" default:"true"`
		EnableCaching     bool `env:"ENABLE_CACHING" default:"true"`
		EnableCompression bool `env:"ENABLE_COMPRESSION" default:"false"`
	} `prefix:"FEATURE_"`

	RateLimiting struct {
		Enabled           bool          `env:"ENABLED" default:"true"`
		RequestsPerSecond int           `env:"RPS" default:"1000" validate:"min=1,max=100000"`
		BurstSize         int           `env:"BURST_SIZE" default:"10000" validate:"min=1,max=1000000"`
		WindowSize        time.Duration `env:"WINDOW_SIZE" default:"1s" validate:"min=100ms,max=1m"`
	} `prefix:"RATELIMIT_"`
}

// ExampleMicroservice shows how to load microservice configuration
// with all the typical dependencies and services.
func ExampleMicroservice() error {
	cfg, err := confkit.Load[MicroserviceConfig](
		confkit.FromEnv(),
		confkit.FromYAMLOptional("microservice-config.yaml"),
		confkit.FromYAMLOptional("defaults.yaml"),
	)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	log.Printf("🚀 Starting service: %s", cfg.Service.Name)
	log.Printf("📊 Database: %s@%s:%d", cfg.PostgreSQL.User, cfg.PostgreSQL.Host, cfg.PostgreSQL.Port)
	log.Printf("💾 Cache: %s:%d db=%d", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Database)
	log.Printf("📬 Message Queue: %s://%s:%d (type=%s)", cfg.MessageQueue.Type, cfg.MessageQueue.Host, cfg.MessageQueue.Port, cfg.MessageQueue.Type)
	log.Printf("📈 Observability: logs=%s, metrics_port=%d, tracing=%v", cfg.Observability.LogLevel, cfg.Observability.MetricsPort, cfg.Observability.TracingEnabled)
	log.Printf("🔐 Auth: JWT enabled, token_expiry=%v", cfg.Auth.TokenExpiry)
	log.Printf("⚡ Rate limiting: %d RPS (burst=%d)", cfg.RateLimiting.RequestsPerSecond, cfg.RateLimiting.BurstSize)

	return nil
}
