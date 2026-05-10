package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/MimoJanra/confkit/aws"
	"github.com/MimoJanra/confkit/k8s"
	"github.com/MimoJanra/confkit/vault"
)

// CloudNativeConfig demonstrates configuration for cloud-native applications
// using Kubernetes ConfigMaps, AWS Secrets Manager, HashiCorp Vault, etc.
type CloudNativeConfig struct {
	App struct {
		Name      string `env:"APP_NAME" validate:"required"`
		Version   string `env:"APP_VERSION"`
		Namespace string `env:"NAMESPACE" default:"default"`
	} `prefix:"APP_"`

	Server struct {
		Port             int           `env:"PORT" default:"8080" validate:"min=1,max=65535"`
		MaxConnections   int           `env:"MAX_CONNECTIONS" default:"1000" validate:"min=1,max=100000"`
		GracefulShutdown time.Duration `env:"GRACEFUL_SHUTDOWN" default:"30s" validate:"min=1s,max=5m"`
		HealthCheckPath  string        `env:"HEALTH_CHECK_PATH" default:"/health"`
		ReadinessPath    string        `env:"READINESS_PATH" default:"/ready"`
		LivenessPath     string        `env:"LIVENESS_PATH" default:"/live"`
	} `prefix:"SERVER_"`

	Database struct {
		Host              string        `env:"HOST" validate:"required"`
		Port              int           `env:"PORT" default:"5432" validate:"min=1,max=65535"`
		User              string        `env:"USER" validate:"required"`
		Password          string        `env:"PASSWORD" validate:"required" secret:"true"`
		Database          string        `env:"DATABASE" validate:"required"`
		SSLMode           string        `env:"SSL_MODE" default:"require" validate:"oneof=disable,allow,prefer,require,verify-ca,verify-full"`
		MaxConnections    int           `env:"MAX_CONNECTIONS" default:"20" validate:"min=1,max=100"`
		ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" default:"30s" validate:"min=5s,max=5m"`
		IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" default:"10m" validate:"min=1m,max=1h"`
	} `prefix:"DB_"`

	Cache struct {
		Enabled    bool          `env:"ENABLED" default:"true"`
		Backend    string        `env:"BACKEND" default:"redis" validate:"oneof=redis,memcached,in-memory"`
		Host       string        `env:"HOST" validate:"required"`
		Port       int           `env:"PORT" default:"6379" validate:"min=1,max=65535"`
		Password   string        `env:"PASSWORD" secret:"true"`
		Database   int           `env:"DATABASE" default:"0" validate:"min=0,max=15"`
		DefaultTTL time.Duration `env:"DEFAULT_TTL" default:"1h" validate:"min=1m,max=24h"`
		MaxMemory  string        `env:"MAX_MEMORY" default:"256mb"`
	} `prefix:"CACHE_"`

	Observability struct {
		LogLevel        string `env:"LOG_LEVEL" default:"info" validate:"oneof=debug,info,warn,error"`
		LogFormat       string `env:"LOG_FORMAT" default:"json" validate:"oneof=json,text"`
		MetricsEnabled  bool   `env:"METRICS_ENABLED" default:"true"`
		MetricsPort     int    `env:"METRICS_PORT" default:"8081" validate:"min=1024,max=65535"`
		TracingEnabled  bool   `env:"TRACING_ENABLED" default:"true"`
		TracingSampler  string `env:"TRACING_SAMPLER" default:"parentbased_always_on"`
		TracingExporter string `env:"TRACING_EXPORTER" default:"otlp" validate:"oneof=otlp,jaeger,zipkin"`
		TraceOTLPURL    string `env:"TRACE_OTLP_URL" default:"http://otel-collector:4317"`
	} `prefix:"OBS_"`

	Security struct {
		EnableTLS          bool     `env:"ENABLE_TLS" default:"true"`
		CertPath           string   `env:"CERT_PATH" validate:"required"`
		KeyPath            string   `env:"KEY_PATH" validate:"required"`
		EnableMTLS         bool     `env:"ENABLE_MTLS" default:"false"`
		CAPath             string   `env:"CA_PATH"`
		CORSEnabled        bool     `env:"CORS_ENABLED" default:"false"`
		CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS"`
	} `prefix:"SEC_"`

	HorizontalScaling struct {
		MinReplicas     int           `env:"MIN_REPLICAS" default:"2" validate:"min=1"`
		MaxReplicas     int           `env:"MAX_REPLICAS" default:"10" validate:"min=1"`
		TargetCPU       int           `env:"TARGET_CPU" default:"80" validate:"min=1,max=100"`
		TargetMemory    int           `env:"TARGET_MEMORY" default:"85" validate:"min=1,max=100"`
		ScaleDownWindow time.Duration `env:"SCALE_DOWN_WINDOW" default:"5m" validate:"min=1m,max=60m"`
		ScaleUpWindow   time.Duration `env:"SCALE_UP_WINDOW" default:"1m" validate:"min=30s,max=10m"`
	} `prefix:"HSCALE_"`

	ResourceLimits struct {
		CPURequest    string `env:"CPU_REQUEST" default:"100m"`
		CPULimit      string `env:"CPU_LIMIT" default:"500m"`
		MemoryRequest string `env:"MEMORY_REQUEST" default:"128Mi"`
		MemoryLimit   string `env:"MEMORY_LIMIT" default:"512Mi"`
	} `prefix:"RESOURCES_"`

	Features struct {
		EnableNewAPI             bool `env:"ENABLE_NEW_API" default:"false"`
		EnableBetaFeatures       bool `env:"ENABLE_BETA_FEATURES" default:"false"`
		EnableMetricsV2          bool `env:"ENABLE_METRICS_V2" default:"false"`
		EnableDistributedTracing bool `env:"ENABLE_DISTRIBUTED_TRACING" default:"true"`
	} `prefix:"FEATURE_"`
}

// ExampleCloudNative shows how to load configuration from multiple cloud sources
// for Kubernetes and other cloud-native deployments.
func ExampleCloudNative() error {
	cfg, err := confkit.Load[CloudNativeConfig](

		confkit.FromEnv(),

		k8s.FromKubernetesConfigMap("default", "app-config"),

		aws.FromAWSSecretsManager("prod/app-secrets"),

		vault.FromVault(
			"https://vault.example.com",
			vault.VaultTokenAuth("hvs.CAESIMgR..."),
			"/secret/myapp",
		),

		confkit.FromYAMLOptional("config.yaml"),

		confkit.FromYAMLOptional("defaults.yaml"),
	)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	log.Printf("🚀 Starting %s v%s in namespace %s", cfg.App.Name, cfg.App.Version, cfg.App.Namespace)
	log.Printf("📡 Server: port=%d, health=%s, readiness=%s, liveness=%s", cfg.Server.Port, cfg.Server.HealthCheckPath, cfg.Server.ReadinessPath, cfg.Server.LivenessPath)
	log.Printf("🗄️  Database: %s@%s:%d (SSL=%s)", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode)
	log.Printf("💾 Cache: %s (%s:%d db=%d)", cfg.Cache.Backend, cfg.Cache.Host, cfg.Cache.Port, cfg.Cache.Database)
	log.Printf("📊 Observability: logs=%s, metrics=%v (port=%d), tracing=%v (%s)", cfg.Observability.LogLevel, cfg.Observability.MetricsEnabled, cfg.Observability.MetricsPort, cfg.Observability.TracingEnabled, cfg.Observability.TracingExporter)
	log.Printf("🔐 Security: TLS=%v, mTLS=%v, CORS=%v", cfg.Security.EnableTLS, cfg.Security.EnableMTLS, cfg.Security.CORSEnabled)
	log.Printf("📈 Scaling: %d-%d replicas, CPU=%d%%, Memory=%d%%", cfg.HorizontalScaling.MinReplicas, cfg.HorizontalScaling.MaxReplicas, cfg.HorizontalScaling.TargetCPU, cfg.HorizontalScaling.TargetMemory)

	return nil
}
