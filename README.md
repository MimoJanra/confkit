# confkit

[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue)](https://golang.org/doc/devel/release)
[![Go Reference](https://pkg.go.dev/badge/github.com/MimoJanra/confkit.svg)](https://pkg.go.dev/github.com/MimoJanra/confkit)
[![Tests](https://github.com/MimoJanra/confkit/actions/workflows/test.yml/badge.svg)](https://github.com/MimoJanra/confkit/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/MimoJanra/confkit/branch/main/graph/badge.svg)](https://codecov.io/gh/MimoJanra/confkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/MimoJanra/confkit)](https://goreportcard.com/report/github.com/MimoJanra/confkit)
[![Documentation](https://img.shields.io/badge/docs-mimojanra.github.io-blue)](https://mimojanra.github.io/confkit/)
[![LLM Context](https://img.shields.io/badge/llms.txt-reference-brightgreen)](./llms.txt)

**Typed, validated configuration loading for Go** — no more stringly-typed config, boilerplate, or cryptic error messages.

confkit lets you define application config as a Go struct, load it from multiple sources (YAML, env vars, JSON, TOML, Kubernetes, AWS, Vault), apply defaults, validate fields, and safely redact secrets. Think Pydantic for Go.

## 30-Second Example

Define your config struct once. Defaults and validation are in the tags:

```go
type Config struct {
    Port     int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    Database string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

// Load from YAML + env, get typed value or clear error
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),  // env overrides YAML
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

On validation error:
```
Invalid configuration:

  Database
    error: field is required
    source: env (DATABASE_URL)
```

No custom error handling. No secret leaks in logs. Types are checked at compile time.

## Why confkit?

✅ **Typed** — `Load[T]` returns your struct, not a map[string]interface{} or error-prone interface{}  
✅ **Defaults & validation** — via struct tags, no extra config files  
✅ **Clear errors** — know exactly which field failed, why, and where it came from  
✅ **Secret redaction** — mark sensitive fields with `secret:"true"`, they're automatically hidden  
✅ **Multiple sources** — load from YAML, env, JSON, TOML, Kubernetes, AWS, Vault with explicit precedence  
✅ **Lightweight** — only 2 core dependencies, cloud integrations are optional modules  
✅ **Production-ready** — v0.5.0 with full test coverage, used in real services

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

## Comparison with Alternatives

confkit is best when you want **typed config structs with built-in validation, defaults, and safe error messages** without assembling multiple libraries by hand.

| | confkit | Viper | envconfig | koanf |
|---|:---:|:---:|:---:|:---:|
| Typed `Load[T]` | ✅ | ❌ | ⚠️ | ❌ |
| Defaults via tags | ✅ | ⚠️ | ✅ | ❌ |
| Validation rules | ✅ | ❌ | ❌ | ❌ |
| Secret redaction | ✅ | ❌ | ❌ | ❌ |
| Multi-source merging | ✅ | ✅ | ⚠️ | ✅ |
| Lightweight core | ✅ | ❌ | ✅ | ✅ |
| Cloud integrations | optional | bundled | bundled | bundled |
| Runtime reloading | ✅ | ✅ | ❌ | ⚠️ |

**confkit shines when:**
- You want a single struct definition for your entire config
- You need defaults and validation without extra code
- You care about safe error messages (no secret leaks)
- You use cloud sources (Vault, AWS) but don't want 50MB of SDKs in your core binary

**Use Viper if:** you need heavy runtime reloading with watches across dozens of files  
**Use envconfig if:** you only care about env vars and simple type conversion  
**Use koanf if:** you want extreme modularity and don't need validation

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

## Real-World Examples

### HTTP Server with Database

```go
type Config struct {
    Server struct {
        Addr string `env:"ADDR" default:":8080"`
        TLS  bool   `env:"TLS" default:"false"`
    }
    Database struct {
        URL      string `env:"URL" validate:"required" secret:"true"`
        MaxConns int    `env:"MAX_CONNS" default:"10"`
    }
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
```

**Environment:**
```bash
SERVER_ADDR=:3000
DATABASE_URL=postgres://user:pass@localhost/db
DATABASE_MAX_CONNS=20
```

### CLI Tool with Multiple Sources

```go
type Config struct {
    Verbose  bool   `flag:"verbose" short:"v"`
    Output   string `flag:"output" short:"o" default:"stdout"`
    InputDir string `flag:"input" validate:"required"`
}

cfg, err := confkit.Load[Config](confkit.FromFlags())
// Use: ./mytool -v -o file.txt --input /data
```

### Microservice with Vault

```go
import "github.com/MimoJanra/confkit/vault"

type Config struct {
    API struct {
        Key    string `validate:"required" secret:"true"`
        Secret string `validate:"required" secret:"true"`
    }
}

auth := vault.VaultTokenAuth(os.Getenv("VAULT_TOKEN"))
cfg, err := confkit.Load[Config](
    vault.FromVault("https://vault.example.com", auth, "/secret/myapp"),
)
```

### Development vs Production

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.defaults.yaml"),      // base defaults
    confkit.FromYAML("config." + os.Getenv("ENV") + ".yaml"), // prod/dev specific
    confkit.FromEnv(),                             // runtime overrides
)
// Loads: config.defaults.yaml → config.prod.yaml → env vars
```

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

MIT — A permissive, redistributable license with minimal restrictions.

confkit is released under the MIT License, which allows you to:
- ✅ Use commercially (SaaS, proprietary software, etc.)
- ✅ Modify and redistribute
- ✅ Use in closed-source projects
- ✅ Sublicense

The only requirement: include a copy of the license in your distribution.

See [LICENSE](LICENSE) file for full text.
