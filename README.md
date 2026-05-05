# confkit

[![Go Version](https://img.shields.io/badge/go-1.24%2B-blue)](https://golang.org/doc/devel/release)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/MimoJanra/confkit.svg)](https://github.com/MimoJanra/confkit/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/MimoJanra/confkit)](https://goreportcard.com/report/github.com/MimoJanra/confkit)
[![Documentation](https://img.shields.io/badge/docs-mimojanra.github.io-blue)](https://mimojanra.github.io/confkit/)

**[📖 Full Documentation](https://mimojanra.github.io/)** — Getting started, API reference, examples, and cloud integrations.

> **Typed, validated configuration loading for Go** — the Pydantic equivalent for Go services.
>
> Load configuration from multiple sources (environment, YAML, JSON, TOML, Kubernetes, AWS, Vault, Consul, etcd) with type safety, validation, and human-readable error messages.

Define your config as a Go struct. Declare sources. Get a fully validated, type-safe value — or a human-readable error message that tells you exactly which field failed, from which source, and why.

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL"        validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
    // Invalid configuration:
    //
    //   DSN
    //     source: env
    //     error: field is required
}
```

**Go 1.24+ · zero mandatory dependencies beyond yaml.v3 and go-toml/v2 · MIT**

---

## Why confkit instead of Viper or koanf

| | confkit | Viper | koanf |
|---|---|---|---|
| Typed return value | ✅ generics `Load[T]` | ❌ `GetString()` | ❌ `Unmarshal()` |
| Validation built-in | ✅ `validate` tag | ❌ manual | ❌ manual |
| Human-readable errors | ✅ field + source + value | ❌ | ❌ |
| Secret redaction | ✅ `secret:"true"` | ❌ | ❌ |
| String interpolation | ✅ `${VAR}` | ❌ | ❌ |
| Core dep footprint | 2 packages | ~20 packages | modular |
| Enterprise sources | optional submodules | bundled | bundled |

**confkit is for you if:** you want your config validated at startup with clear error messages, you don't want stringly-typed `Get*()` accessors, and you don't want cloud SDKs downloaded unless you use them.

---

## Install

```bash
go get github.com/MimoJanra/confkit
```

Enterprise sources (Vault, Consul, etcd, AWS) are separate optional modules:

```bash
go get github.com/MimoJanra/confkit/vault
go get github.com/MimoJanra/confkit/consul
go get github.com/MimoJanra/confkit/etcd
go get github.com/MimoJanra/confkit/aws
```

---

## Quick Start

```go
package main

import (
    "log"
    "time"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Host    string        `env:"HOST"     default:"localhost"`
    Port    int           `env:"PORT"     default:"8080"      validate:"min=1,max=65535"`
    Timeout time.Duration `env:"TIMEOUT"  default:"30s"`
    DB      struct {
        DSN      string `env:"DSN"       validate:"required" secret:"true"`
        MaxConns int    `env:"MAX_CONNS" default:"10"        validate:"min=1,max=100"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromYAML("config.yaml"), // lowest priority
        confkit.FromEnv(),               // overrides file
        confkit.FromFlags(),             // highest priority
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    log.Printf("listening on %s:%d", cfg.Host, cfg.Port)
}
```

Environment variables for the struct above: `HOST`, `PORT`, `TIMEOUT`, `DB_DSN`, `DB_MAX_CONNS`.

---

## Struct Tags Reference

```
env:"VAR_NAME"          — read from environment variable VAR_NAME
flag:"flag-name"        — read from CLI flag --flag-name
short:"f"               — single-character short flag -f
default:"value"         — use this value when no source provides one
validate:"rules"        — validation rules (see Validation section)
secret:"true"           — redact this field in errors, dumps, and logs
prefix:"PREFIX_"        — prepend to env names of all fields in a nested struct
help:"description"      — description shown in schema and CLI help
hidden:"true"           — hide from CLI help output
```

---

## Sources

Built-in sources — pass any combination, last one wins per field:

```go
confkit.FromYAML(path string) Source
confkit.FromJSON(path string) Source
confkit.FromTOML(path string) Source
confkit.FromEnv() Source
confkit.FromFlags() Source
confkit.FromKubernetesConfigMap(namespace, name string) Source
confkit.FromKubernetesConfigMapWithPath(namespace, name, mountPath string) Source
```

Optional enterprise sources (separate `go get` per module):

```go
// go get confkit/vault
vault.FromVault(addr string, auth VaultAuth, pathPrefix string) confkit.Source
vault.FromVaultWithKVVersion(addr string, auth VaultAuth, kvVersion int, pathPrefix string) confkit.Source

// go get confkit/consul
consul.FromConsul(addr string) confkit.Source
consul.FromConsulWithToken(addr, token string) confkit.Source
consul.FromConsulWithOptions(addr, token, datacenter string) confkit.Source

// go get confkit/etcd
etcd.FromEtcd(endpoints []string) confkit.Source
etcd.FromEtcdWithPrefix(endpoints []string, prefix string) confkit.Source

// go get confkit/aws
aws.FromAWSSSMParameterStore(pathPrefix string) confkit.Source
aws.FromAWSSSMParameterStoreWithTTL(pathPrefix string, ttl time.Duration) confkit.Source
aws.FromAWSSecretsManager(secretName string) confkit.Source
aws.FromAWSSecretsManagerWithRegion(secretName, region string) confkit.Source
aws.FromAWSSecretsManagerWithOptions(secretName, region string, ttl time.Duration) confkit.Source
aws.FromAWSSecretsManagerMultiRegion(secretName string, regions []string) confkit.Source
aws.FromAWSSSMParameterStoreMultiRegion(pathPrefix string, regions []string) confkit.Source
```

---

## Validation

Rules used in the `validate` struct tag:

| Rule | Types | Behaviour |
|---|---|---|
| `required` | any | Non-zero value required |
| `min=N` | int, float | Value ≥ N |
| `min=N` | string | Length ≥ N characters |
| `max=N` | int, float | Value ≤ N |
| `max=N` | string | Length ≤ N characters |
| `oneof=a b c` | string | Value must equal one of the space-separated options |

Rules are comma-separated: `validate:"required,min=1,max=65535"`.

```go
type Config struct {
    Port     int    `env:"PORT"      validate:"required,min=1,max=65535"`
    LogLevel string `env:"LOG_LEVEL" validate:"required,oneof=debug info warn error" default:"info"`
    Name     string `env:"APP_NAME"  validate:"required,min=3,max=64"`
}
```

### Custom validator

Register a named validator per load-call — no global state:

```go
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("port-range", func(v reflect.Value) error {
        n := v.Int()
        if n < 1024 || n > 49151 {
            return fmt.Errorf("must be a registered port (1024–49151)")
        }
        return nil
    }),
)
// use in tag: validate:"port-range"
```

---

## Load API

```go
// Load using variadic sources — shorthand for the common case
func Load[T any](sources ...Source) (T, error)

// Load with fine-grained options (validators, middleware, interpolation depth)
func LoadWithOptions[T any](options ...Option) (T, error)

// Load and set up a file watcher in one call
func LoadWithWatcher[T any](filePath string, sources ...Source) (T, *ConfigWatcher, error)

// Option constructors
func WithSource(source Source) Option
func WithValidator(name string, fn func(reflect.Value) error) Option
func WithMiddleware(fn MiddlewareFunc) Option
func WithInterpolationMaxDepth(depth int) Option
```

---

## Error Handling

```go
// Explain formats any confkit error into a human-readable multi-line string.
// Returns err.Error() for non-confkit errors, "" for nil.
func Explain(err error) string

// ErrorReport is returned as error from Load. It implements error.
type ErrorReport struct {
    Errors []FieldError
}

type FieldError struct {
    Path    string    // "Database.Password"
    Source  string    // "env", "yaml", "validation"
    Kind    ErrorKind // parse | validation | missing | io
    Rule    string    // "required", "min", ...
    Message string
    Value   string    // empty if Secret == true
    Secret  bool
}
```

---

## Secrets

Fields tagged `secret:"true"` are redacted everywhere:

```go
type Config struct {
    Token    string `env:"API_TOKEN"  secret:"true" validate:"required"`
    Password string `env:"DB_PASSWORD" secret:"true"`
}
```

- Error messages show `<redacted>` instead of the value
- `DumpConfig` substitutes `"***REDACTED***"`
- Safe to log the output of `Explain(err)` and `DumpConfig` without leaking credentials

```go
fields := confkit.ScanFields(cfg)
data, _ := confkit.DumpConfig(cfg, fields)
// {"DB.Password": "***REDACTED***", "Host": "localhost", ...}
```

---

## String Interpolation

Values can reference other fields or env vars using `${NAME}`:

```go
type Config struct {
    Host    string `env:"HOST"     default:"localhost"`
    Port    int    `env:"PORT"     default:"8080"`
    BaseURL string `env:"BASE_URL" default:"http://${HOST}:${PORT}/api"`
}
// BaseURL → "http://localhost:8080/api" unless overridden
```

Resolution order: config fields first, then OS environment. Circular references are detected and returned as errors.

---

## Nested Structs and Prefixes

```go
type Config struct {
    App  AppConfig
    DB   DBConfig   `prefix:"DB_"`
    Cache CacheConfig `prefix:"CACHE_"`
}

type DBConfig struct {
    Host     string `env:"HOST"     default:"localhost"`
    Port     int    `env:"PORT"     default:"5432"`
    Password string `env:"PASSWORD" secret:"true" validate:"required"`
}
// Reads DB_HOST, DB_PORT, DB_PASSWORD from env
```

Nesting is unlimited. Prefixes from all ancestor structs are concatenated.

---

## Hot Reload

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml",
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}

watcher.AddListener(func(oldCfg, newCfg any, err error) {
    if err != nil {
        log.Printf("reload failed: %v", err)
        return
    }
    log.Println("config reloaded")
})

watcher.SetPollInterval(5 * time.Second) // default: 500ms
watcher.Start()
defer watcher.Stop()
```

The watcher polls `mtime`. When the file changes, all listeners are called with the old and new config cast to `any`.

---

## Custom Source

```go
// The complete Source interface — implement these two methods:
type Source interface {
    Name() string
    Lookup(field *FieldInfo) (any, bool, error)
    // return ("", false, nil)  — field not found in this source
    // return (value, true, nil) — found
    // return ("", false, err)  — source error
}

// FieldInfo fields available to your Lookup implementation:
// .Name         string            — "Password"
// .Path         string            — "Database.Password"
// .Tags         map[string]string — all struct tags
// .IsSecret     bool
// .HasDefault   bool
// .AncestorTags []map[string]string — tags of parent structs

// Helper for returning a permanently-errored source (e.g. from a constructor):
func NewErrorSource(err error) Source
```

---

## Schema Generation

```go
import "github.com/MimoJanra/confkit/schema"

// JSON Schema (draft-07 compatible)
s, err := schema.GenerateSchema[Config]()
data, _ := json.MarshalIndent(s, "", "  ")

// Markdown reference table
md := schema.GenerateMarkdown[Config]()

// CLI --help style output
help := schema.GenerateCLIHelp[Config]()
```

---

## Supported Types

| Type | Parsed from |
|---|---|
| `string` | as-is |
| `int` / `int8` / `int16` / `int32` / `int64` | decimal |
| `uint` / `uint8` / `uint16` / `uint32` / `uint64` | decimal |
| `float32` / `float64` | decimal |
| `bool` | `true` `false` `1` `0` `yes` `no` |
| `time.Duration` | `"5s"` `"1m30s"` `"2h"` |
| `[]string` | comma-separated `"a,b,c"` |
| `[]int` | comma-separated `"1,2,3"` |

---

## Vault Auth Methods

```go
vault.VaultTokenAuth(token string) VaultAuth
vault.VaultAppRoleAuth(roleID, secretID string) VaultAuth
vault.VaultKubernetesAuth(role, jwt string) VaultAuth
```

---

## Documentation

- **[Full Documentation](https://mimojanra.github.io/)** — Getting started, guides, API reference
- **[Getting Started](https://mimojanra.github.io/docs/getting-started/)** — 5-minute quick start
- **[API Reference](https://mimojanra.github.io/api/)** — Complete function and type reference
- **[Examples](https://mimojanra.github.io/examples/)** — Runnable code examples
- **[Sources Guide](https://mimojanra.github.io/docs/sources/)** — All configuration sources
- **[GitHub Discussions](https://github.com/MimoJanra/confkit/discussions)** — Questions and ideas
- **[Issues](https://github.com/MimoJanra/confkit/issues)** — Bug reports and feature requests

---

## License

MIT
