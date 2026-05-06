// Package confkit provides typed, validated configuration loading for Go services.
//
// Load configuration from multiple sources (YAML, JSON, TOML, environment variables,
// Kubernetes ConfigMaps, AWS SSM, HashiCorp Vault, Consul, etcd) with type safety,
// validation, and human-readable error messages.
//
// # Quick Start
//
// Define your configuration as a Go struct with tags:
//
//	type Config struct {
//	    Port        int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
//	    DatabaseURL string `env:"DATABASE_URL" validate:"required" secret:"true"`
//	}
//
// Load from sources with explicit precedence:
//
//	cfg, err := confkit.Load[Config](
//	    confkit.FromYAML("config.yaml"),
//	    confkit.FromEnv(),
//	)
//	if err != nil {
//	    log.Fatal(confkit.Explain(err))
//	}
//
// # Features
//
// • Type-safe configuration via Load[T] generics
// • Validation at startup (fail fast with clear errors)
// • Automatic secret redaction (secret:"true" tags)
// • Multiple sources with explicit precedence
// • String interpolation (${VAR} expansion)
// • Hot reload support for config changes
// • Optional enterprise integrations (Vault, Consul, etcd, AWS)
//
// # Sources
//
// Built-in sources for common formats:
//
//	confkit.FromYAML(path)          // Load from YAML file
//	confkit.FromJSON(path)          // Load from JSON file
//	confkit.FromTOML(path)          // Load from TOML file
//	confkit.FromEnv()               // Load from environment variables
//	confkit.FromFlags()             // Load from command-line flags
//	k8s.FromKubernetesConfigMap(namespace, name)      // Load from K8s ConfigMap (github.com/MimoJanra/confkit/k8s)
//
// Optional enterprise sources (separate modules):
//
//	vault.FromVault(addr, auth, pathPrefix)        // HashiCorp Vault
//	consul.FromConsul(addr)                        // Consul KV
//	etcd.FromEtcd(endpoints)                       // etcd v3
//	aws.FromAWSSSMParameterStore(pathPrefix)       // AWS Systems Manager
//	aws.FromAWSSecretsManager(secretName)          // AWS Secrets Manager
//
// # Struct Tags
//
// Use struct tags to control how fields are loaded:
//
//	env:"VAR"         Read from environment variable VAR
//	yaml:"field"      Read from YAML field 'field'
//	default:"value"   Use this value if no source provides one
//	validate:"rules"  Apply validation rules (required, min, max, oneof)
//	secret:"true"     Redact this field in all errors and logs
//	prefix:"PRE_"     Add PRE_ prefix to env vars in nested struct
//	help:"text"       Description for schema and help text
//
// # Validation
//
// Built-in validation rules:
//
//	required          Non-zero value required
//	min=N             Minimum value (int, float) or length (string)
//	max=N             Maximum value (int, float) or length (string)
//	oneof=a b c       Value must equal one of the options
//
// Example:
//
//	type Config struct {
//	    Port     int    `env:"PORT" validate:"required,min=1,max=65535"`
//	    LogLevel string `env:"LOG_LEVEL" validate:"oneof=debug info warn error" default:"info"`
//	}
//
// # Secret Redaction
//
// Fields tagged with secret:"true" are automatically redacted in error messages,
// config dumps, and logs:
//
//	type Config struct {
//	    APIKey   string `env:"API_KEY" secret:"true"`
//	    Password string `env:"PASSWORD" secret:"true"`
//	}
//
// When loading fails, secrets appear as "***" instead of actual values.
//
// # Error Handling
//
// Use Explain() to get human-readable error messages:
//
//	cfg, err := confkit.Load[Config](confkit.FromEnv())
//	if err != nil {
//	    fmt.Println(confkit.Explain(err))
//	    // Output:
//	    // Invalid configuration:
//	    //
//	    //   DatabaseURL
//	    //     error: field is required
//	    //     source: env (DATABASE_URL)
//	}
//
// # Nested Structures
//
// Use nested structs with optional env prefix:
//
//	type Config struct {
//	    Server struct {
//	        Host string `env:"HOST"`
//	        Port int    `env:"PORT"`
//	    }
//	    Database struct {
//	        URL string `env:"URL"`
//	    } `prefix:"DB_"`
//	}
//
// Reads from: SERVER_HOST, SERVER_PORT, DB_URL
//
// # Hot Reload
//
// LoadWithWatcher supports hot reload when files change:
//
//	cfg, watcher, err := confkit.LoadWithWatcher[Config](
//	    "config.yaml",
//	    confkit.FromYAML("config.yaml"),
//	)
//	go func() {
//	    for newCfg := range watcher.Changes() {
//	        log.Printf("Config updated: %+v\n", newCfg)
//	    }
//	}()
//
// # Why confkit?
//
// confkit is for Go services that need:
//
// • Typed config (no stringly-typed Get*() accessors)
// • Validation at startup (fail fast)
// • Clear error messages (human-readable)
// • Secret redaction (never expose credentials)
// • Multiple sources (YAML + env + flags)
// • Optional cloud integrations (without bloating core)
//
// Compare to Viper (flexible but stringly-typed) and koanf (modular but manual validation).
//
// # Documentation
//
// Full documentation is available at: https://mimojanra.github.io/confkit/
//
// See README.md for more examples, or visit pkg.go.dev for this package's API docs.
package confkit
