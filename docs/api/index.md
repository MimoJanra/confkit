---
layout: default
title: API Reference
---

# API Reference

Complete reference for confkit's public API.

## Core Functions

### Load

```go
func Load[T any](sources ...Source) (*T, error)
```

Loads configuration from the provided sources and returns a fully typed, validated config or an error.

**Parameters:**
- `sources` — Variable number of `Source` instances to load from

**Returns:**
- `T` — The configuration struct (zero value on error)
- `error` — An `*ErrorReport` if validation fails

**Example:**
```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}

cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

---

### LoadWithOptions

```go
func LoadWithOptions[T any](options ...Option) (*T, error)
```

Advanced loading with functional options for custom validators, middleware, and interpolation configuration.

**Parameters:**
- `options` — Variable number of `Option` instances

**Returns:**
- `T` — The configuration struct (zero value on error)
- `error` — An `*ErrorReport` if validation fails

**Options:**
- `WithSource(source Source)` — Add a source
- `WithValidator(name string, fn CustomValidatorFunc)` — Register a custom per-field validator
- `WithModelValidator[T](fn func(*T) error)` — Register a cross-field validator
- `WithMiddleware(fn MiddlewareFunc)` — Add a value transformation middleware
- `WithInterpolationMaxDepth(n int)` — Set interpolation recursion limit
- `WithAuditLogger(fn AuditLogger)` — Receive a log of every field's resolved value and source
- `WithLoadHook(fn LoadHookFunc)` — Receive success/duration/errCount after every load

**Example:**
```go
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromYAML("config.yaml")),
    confkit.WithSource(confkit.FromEnv()),
    // Cross-field validation
    confkit.WithModelValidator(func(cfg *Config) error {
        if cfg.TLSEnabled && cfg.CertPath == "" {
            return fmt.Errorf("cert_path required when tls_enabled is true")
        }
        return nil
    }),
    // Audit trail
    confkit.WithAuditLogger(func(entries []confkit.AuditEntry) {
        for _, e := range entries {
            log.Printf("field=%s source=%s value=%s", e.Field, e.Source, e.Value)
        }
    }),
)
```

---

### LoadWithWatcher

```go
func LoadWithWatcher[T any](filePath string, sources ...Source) (*T, *ConfigWatcher, error)
```

Loads configuration and returns a watcher for hot-reloading when files change.

**Parameters:**
- `filePath` — Path to watch for changes
- `sources` — Configuration sources

**Returns:**
- `T` — Initial configuration struct
- `*ConfigWatcher` — Watcher instance for subscribing to changes
- `error` — Load or watcher creation error

**Example:**
```go
cfg, watcher, err := confkit.LoadWithWatcher[Config](
    "config.yaml",
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(err)
}

// Subscribe to changes
go func() {
    for newCfg := range watcher.Changes() {
        log.Printf("Config reloaded: %+v\n", newCfg)
    }
}()
```

---

### Explain

```go
func Explain(err error) string
```

Formats an error report into a human-readable, colorized error message suitable for logging or stdout.

**Parameters:**
- `err` — An error (typically `*ErrorReport`)

**Returns:**
- `string` — Formatted error message

**Example:**
```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    fmt.Println(confkit.Explain(err))
    // Invalid configuration:
    // 
    //   PORT
    //     error: must be between 1 and 65535
    //     got: 99999
    //     source: env (PORT)
}
```

---

## Sources

### FromEnv

```go
func FromEnv() Source
```

Loads configuration from environment variables. Field names are uppercased and matched to `env` tags.

**Example:**
```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" validate:"required"`
    Port int `env:"PORT" default:"8080"`
}

// Reads from $DATABASE_URL and $PORT
cfg, _ := confkit.Load[Config](confkit.FromEnv())
```

---

### FromYAML

```go
func FromYAML(path string) Source
```

Loads configuration from a YAML file.

**Example:**
```go
// config.yaml:
// databaseURL: postgres://localhost/db
// port: 3000

cfg, _ := confkit.Load[Config](confkit.FromYAML("config.yaml"))
```

---

### FromJSON

```go
func FromJSON(path string) Source
```

Loads configuration from a JSON file.

**Example:**
```go
cfg, _ := confkit.Load[Config](confkit.FromJSON("config.json"))
```

---

### FromTOML

```go
func FromTOML(path string) Source
```

Loads configuration from a TOML file.

**Example:**
```go
cfg, _ := confkit.Load[Config](confkit.FromTOML("config.toml"))
```

---

### FromYAMLFiles / FromJSONFiles / FromTOMLFiles

```go
func FromYAMLFiles(paths ...string) Source
func FromJSONFiles(paths ...string) Source
func FromTOMLFiles(paths ...string) Source
```

Merges multiple files of the same format into a single source. First file to provide a value wins; later files fill in only unset fields. Nested maps are deep-merged.

**Example:**
```go
cfg, _ := confkit.Load[Config](
    confkit.FromYAMLFiles("base.yaml", "production.yaml", "local.yaml"),
    confkit.FromEnv(),
)
```

---

### FromFlags

```go
func FromFlags() Source
```

Loads configuration from command-line flags. Field names are converted to kebab-case flags.

Supported flag formats:
- `--key=value` — long flag with `=` separator
- `--key value` — long flag with space separator
- `-k value` — short flag with space separator
- `-k=value` — short flag with `=` separator
- `--flag` — boolean flag (sets field to `"true"`)

**Example:**
```go
// Run: go run main.go --database-url=... --port=3000 --verbose
cfg, _ := confkit.Load[Config](confkit.FromFlags())
```

---

## Observability Submodules

### confkit/prometheus

```go
import "github.com/MimoJanra/confkit/prometheus"

m := prometheus.NewMetrics(prometheus.DefaultRegisterer)

cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    m.Hook(), // records loads_total, load_duration_seconds, errors_total
)
```

**Metrics registered:**
- `confkit_loads_total{status="success|error"}` — counter
- `confkit_load_duration_seconds` — histogram
- `confkit_errors_total{kind="validation"}` — counter

---

### confkit/otel

```go
import "github.com/MimoJanra/confkit/otel"

cfg, err := otel.Load[Config](ctx, tracer,
    confkit.FromEnv(),
    confkit.FromYAML("config.yaml"),
)
// Creates span "confkit.Load" with attributes: confkit.sources, confkit.success
```

---

## Cloud & Enterprise Sources

### FromK8sConfigMap

```go
func FromK8sConfigMap(namespace, configMapName string) confkit.Source
```

Loads configuration from a Kubernetes ConfigMap.

---

### FromAWSSSM

```go
func FromAWSSSM(parameterPath string) confkit.Source
```

Loads configuration from AWS Systems Manager Parameter Store.

---

### FromVault

```go
func FromVault(addr, secretPath string) confkit.Source
```

Loads secrets from HashiCorp Vault.

---

### FromConsul

```go
func FromConsul(addr string) confkit.Source
```

Loads configuration from Consul KV.

---

### FromEtcd

```go
func FromEtcd(endpoints []string) confkit.Source
```

Loads configuration from etcd v3.

---

### FromAWSSecretsManager

```go
func FromAWSSecretsManager(secretName string) confkit.Source
```

Loads secrets from AWS Secrets Manager.

---

## Tags

Configuration is driven by struct tags. Common tags:

| Tag | Purpose | Example |
|-----|---------|---------|
| `env` | Environment variable name | `env:"DATABASE_URL"` |
| `yaml` / `json` | File field name | `yaml:"database_url"` |
| `default` | Default value if not found | `default:"localhost"` |
| `validate` | Validation rules | `validate:"required,min=1,max=65535"` |
| `secret` | Mark as sensitive (redacted) | `secret:"true"` |
| `desc` | Help text / description | `desc:"Database connection URL"` |
| `flag` | CLI long flag name | `flag:"database-url"` |
| `short` | CLI short flag (single char) | `short:"d"` |

### Validation Rules

**Numeric / range:**
- `required` — field must be present and non-zero
- `min=N` — minimum value (int/float) or minimum length (string)
- `max=N` — maximum value (int/float) or maximum length (string)
- `oneof=val1,val2,...` — must be one of the listed values

**Format (string fields):**
- `email` — valid email address
- `url` — valid URL (any scheme)
- `http_url` — valid HTTP or HTTPS URL
- `ip` — valid IPv4 or IPv6 address
- `ipv4` — valid IPv4 address
- `ipv6` — valid IPv6 address
- `uuid` — valid UUID (v1–v5)
- `hostname` — valid hostname per RFC 1123
- `port` — valid port number 1–65535 (works on `int` and `string`)
- `regex=pattern` — must match the regular expression
- `len=N` — must be exactly N characters (Unicode-aware)
- `contains=str` — must contain the substring
- `startswith=str` — must start with the prefix
- `endswith=str` — must end with the suffix
- `alpha` — letters only
- `alphanum` — letters and digits only
- `numeric` — digits only
- `lowercase` — must be all lowercase
- `uppercase` — must be all uppercase
- `notempty` — must not be blank (non-whitespace)

**Example:**
```go
type Config struct {
    Port        int    `env:"PORT"    validate:"required,port"`
    AdminEmail  string `env:"EMAIL"   validate:"required,email"`
    APIKey      string `env:"API_KEY" validate:"required,len=32" secret:"true"`
    Environment string `env:"ENV"     validate:"oneof=dev,staging,prod"`
    ServiceURL  string `env:"SVC_URL" validate:"http_url"`
    ServiceID   string `env:"SVC_ID"  validate:"uuid"`
}
```

---

## Types

### ErrorReport

The error type returned by `Load()`. Implements `error` interface.

```go
type ErrorReport struct {
    Errors []FieldError
}

func (e *ErrorReport) Error() string  // Implements error interface
```

**Methods:**
- `AddError(err FieldError)` — Add a field error
- `IsEmpty() bool` — Check if there are no errors
- `Error() string` — Return a formatted error string (implements `error`)
- `Unwrap() []error` — Return individual field errors as a slice (for use with `errors.As`/`errors.Is`)
- `Format() string` — Return a human-readable, multi-line error report

---

### FieldError

Represents a single field validation or loading error.

```go
type FieldError struct {
    Path    string
    Source  string
    Kind    ErrorKind  // parse | validation | io
    Rule    string
    Message string
    Value   string
    Secret  bool
}
```

---

### Source

Interface for configuration sources. Implement to create custom sources.

```go
type Source interface {
    Name() string
    Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)
}
```

**Example Custom Source:**
```go
type MySource struct{}

func (s *MySource) Name() string {
    return "custom"
}

func (s *MySource) Lookup(ctx context.Context, field *FieldInfo) (any, bool, error) {
    // Return value, found, error
    return "value", true, nil
}

cfg, _ := confkit.Load[Config](confkit.WithSource(&MySource{}))
```

---

## Configuration Struct Tags

### Complete Example

```go
type DatabaseConfig struct {
    Host     string `env:"DB_HOST" default:"localhost"`
    Port     int    `env:"DB_PORT" validate:"min=1,max=65535" default:"5432"`
    Username string `env:"DB_USER" validate:"required"`
    Password string `env:"DB_PASSWORD" validate:"required" secret:"true"`
}

type Config struct {
    Database DatabaseConfig
    LogLevel string `env:"LOG_LEVEL" validate:"oneof=debug,info,warn,error" default:"info"`
}
```

---

## Best Practices

1. **Mark secrets with `secret:"true"`** — Automatically redacted in errors and logs
2. **Use `validate:"required"`** — Catch missing config early
3. **Provide sensible defaults** — Reduce configuration burden
4. **Source order matters** — first source wins; list highest-priority sources first
5. **Use custom validators for complex logic** — Keep validation rules simple and composable

---

## See Also

- [Getting Started](/docs/getting-started/) — Quick start guide
- [Validation](/docs/validation/) — Deep dive into validation
- [Error Handling](/docs/errors/) — Programmatic error handling
