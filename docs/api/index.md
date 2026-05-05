---
layout: default
title: API Reference
---

# API Reference

Complete reference for confkit's public API.

## Core Functions

### Load

```go
func Load[T any](sources ...Source) (T, error)
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
func LoadWithOptions[T any](options ...Option) (T, error)
```

Advanced loading with functional options for custom validators, middleware, and interpolation configuration.

**Parameters:**
- `options` — Variable number of `Option` instances

**Returns:**
- `T` — The configuration struct (zero value on error)
- `error` — An `*ErrorReport` if validation fails

**Options:**
- `WithSource(source Source)` — Add a source
- `WithValidator(name string, fn CustomValidatorFunc)` — Register custom validator
- `WithMiddleware(fn MiddlewareFunc)` — Add transformation middleware
- `WithInterpolationMax(max int)` — Set interpolation recursion limit

**Example:**
```go
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromYAML("config.yaml")),
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("custom", func(val string) error {
        if val == "invalid" {
            return fmt.Errorf("cannot be 'invalid'")
        }
        return nil
    }),
)
```

---

### LoadWithWatcher

```go
func LoadWithWatcher[T any](filePath string, sources ...Source) (T, *ConfigWatcher, error)
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

### FromFlags

```go
func FromFlags() Source
```

Loads configuration from command-line flags. Field names are converted to kebab-case flags.

**Example:**
```go
// Run: go run main.go --database-url=... --port=3000
cfg, _ := confkit.Load[Config](confkit.FromFlags())
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

### Validation Rules

- `required` — Field must be present
- `min=N` — Minimum value (int, float, string length)
- `max=N` — Maximum value (int, float, string length)
- `oneof=val1,val2,...` — Must be one of the listed values

**Example:**
```go
type Config struct {
    Port int `env:"PORT" validate:"required,min=1,max=65535"`
    Status string `validate:"oneof=active,inactive,pending"`
    DatabaseURL string `validate:"required" secret:"true"`
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
- `HasErrors() bool` — Check if there are errors
- `FieldErrors() []FieldError` — Get all errors

---

### FieldError

Represents a single field validation or loading error.

```go
type FieldError struct {
    Path    string
    Kind    ErrorKind
    Message string
    Value   string
    Source  string
}
```

---

### Source

Interface for configuration sources. Implement to create custom sources.

```go
type Source interface {
    Name() string
    Lookup(field *FieldInfo) (Value, bool, error)
}
```

**Example Custom Source:**
```go
type MySource struct{}

func (s *MySource) Name() string {
    return "custom"
}

func (s *MySource) Lookup(field *FieldInfo) (confkit.Value, bool, error) {
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
4. **Source order matters** — Left-to-right precedence (later sources override earlier)
5. **Use custom validators for complex logic** — Keep validation rules simple and composable

---

## See Also

- [Getting Started](/docs/getting-started/) — Quick start guide
- [Validation](/docs/validation/) — Deep dive into validation
- [Error Handling](/docs/errors/) — Programmatic error handling
