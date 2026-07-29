# confkit

[![Go Version](https://img.shields.io/badge/go-1.26%2B-blue)](https://golang.org/doc/devel/release)
[![Go Reference](https://pkg.go.dev/badge/github.com/MimoJanra/confkit.svg)](https://pkg.go.dev/github.com/MimoJanra/confkit)
[![Tests](https://github.com/MimoJanra/confkit/actions/workflows/test.yml/badge.svg)](https://github.com/MimoJanra/confkit/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/MimoJanra/confkit/branch/main/graph/badge.svg)](https://codecov.io/gh/MimoJanra/confkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/MimoJanra/confkit)](https://goreportcard.com/report/github.com/MimoJanra/confkit)
[![Documentation](https://img.shields.io/badge/docs-confkit.xyz-blue)](https://confkit.xyz/)
[![LLM Context](https://img.shields.io/badge/llms.txt-reference-brightgreen)](./llms.txt)

**Typed, validated configuration loading for Go** — no more stringly-typed config, boilerplate, or cryptic error
messages.

confkit lets you define application config as a Go struct, load it from multiple sources (YAML, env vars, JSON, TOML,
Kubernetes, AWS, Vault), apply defaults, validate fields, and safely redact secrets. Think Pydantic for Go.

## 30-Second Example

Define your config struct once. Defaults and validation are in the tags:

```go
type Config struct {
    Port     int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    Database string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

// Load from env + YAML, get typed value or clear error
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),           // env wins — checked first
    confkit.FromYAML("config.yaml"),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

On validation error — note the last line:

```
Invalid configuration:

  Database
    source: validation
    error: field is required
    got: ***REDACTED***
```

Every problem is reported at once, not just the first. No custom error handling. No secret leaks in logs — the
field is tagged `secret:"true"`, so its value is redacted here and in dumps and audit logs too.

## Why confkit?

✅ **Typed** — `Load[T]` returns your struct, not a map[string]interface{} or error-prone interface{}  
✅ **Defaults & validation** — via struct tags, no extra config files  
✅ **Clear errors** — know exactly which field failed, why, and where it came from  
✅ **Secret redaction** — mark sensitive fields with `secret:"true"`, they're automatically hidden  
✅ **Multiple sources** — load from YAML, env, JSON, TOML, Kubernetes, AWS, Vault with explicit precedence  
✅ **Lightweight** — only 2 core dependencies, cloud integrations are optional modules  
✅ **Production-ready** — v1.0.3, 85% core coverage enforced in CI, documented API stability guarantee

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL"        validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    confkit.FromYAML("config.yaml"),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
    // Invalid configuration:
    //
    //   DSN
    //     source: validation
    //     error: field is required
    //     got: ***REDACTED***
}
```

**Go 1.26.5+ · zero mandatory dependencies beyond yaml.v3 and go-toml/v2 · MIT**

---

## Comparison with Alternatives

confkit is best when you want **typed config structs with built-in validation, defaults, and safe error messages**
without assembling multiple libraries by hand.

|                      | confkit  |  Viper  | envconfig |  koanf  |
|----------------------|:--------:|:-------:|:---------:|:-------:|
| Typed `Load[T]`      |    ✅     |    ❌    |    ⚠️     |    ❌    |
| Defaults via tags    |    ✅     |   ⚠️    |     ✅     |    ❌    |
| Validation rules     |    ✅     |    ❌    |     ❌     |    ❌    |
| Secret redaction     |    ✅     |    ❌    |     ❌     |    ❌    |
| Multi-source merging |    ✅     |    ✅    |    ⚠️     |    ✅    |
| Lightweight core     |    ✅     |    ❌    |     ✅     |    ✅    |
| Cloud integrations   | optional | optional¹ |   N/A    | optional |
| Runtime reloading    |    ✅     |    ✅    |     ❌     |   ⚠️    |

¹ Viper remote requires `import _ "github.com/spf13/viper/remote"` and covers only etcd/Consul/Firebase. confkit cloud modules cover Vault, AWS SSM/Secrets Manager, Consul, etcd, and Kubernetes — each as a separate `go get`.

**confkit shines when:**

- You want a single struct definition for your entire config
- You need defaults and validation without extra code
- You care about safe error messages (no secret leaks)
- You use cloud sources (Vault, AWS) but don't want 50MB of SDKs in your core binary

**Use Viper if:** you need heavy runtime reloading with watches across dozens of files  
**Use envconfig if:** you only care about env vars and simple type conversion  
**Use koanf if:** you want extreme modularity and don't need validation

For detailed comparisons, see:
- **[confkit vs Viper](./docs/comparison/viper.md)**
- **[confkit vs envconfig](./docs/comparison/envconfig.md)**
- **[confkit vs koanf](./docs/comparison/koanf.md)**

---

## Use Cases

confkit is useful for:

- **Go services** that need typed configuration without boilerplate
- **CLI tools** that combine flags, files, and environment variables
- **Kubernetes applications** using ConfigMaps and Secrets
- **Microservices** that load secrets from Vault, Consul, etcd, or AWS
- **Projects** looking for a type-safe alternative to manual env parsing
- **Teams** that want config validation before application startup

---

## Install

```bash
go get github.com/MimoJanra/confkit@latest
```

Enterprise sources (Vault, Consul, etcd, AWS) are separate optional modules:

```bash
go get github.com/MimoJanra/confkit/vault@latest
go get github.com/MimoJanra/confkit/consul@latest
go get github.com/MimoJanra/confkit/etcd@latest
go get github.com/MimoJanra/confkit/aws@latest
```

---

## Breaking Changes in v1.0.0

**`Load[T]` now returns `(*T, error)` instead of `(T, error)` (pointer)**

This aligns with Go idioms for larger configs. Go automatically dereferences pointers for field access, so most code works unchanged:

```go
// v1.0+ — works exactly the same
cfg, err := confkit.Load[Config](confkit.FromEnv())
log.Printf("Port: %d", cfg.Port)  // auto-dereference, works fine
```

If you explicitly need a value type:
```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
cfgValue := *cfg  // explicit dereference if needed
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
		confkit.FromFlags(),             // highest priority — checked first
		confkit.FromEnv(),               // overrides file
		confkit.FromYAML("config.yaml"), // fallback
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
    confkit.FromEnv(),                                         // runtime overrides — highest priority
    confkit.FromYAML("config." + os.Getenv("ENV") + ".yaml"), // prod/dev specific
    confkit.FromYAML("config.defaults.yaml"),                  // base defaults — fallback
)
// Loads: env vars → config.prod.yaml → config.defaults.yaml
```

### Complete Production Examples

See **[`examples/`](./examples/)** directory for fully-working, tested examples:

- **Web Service** — Database, cache, logging configuration
- **Microservice** — PostgreSQL, Redis, RabbitMQ, observability, JWT
- **CLI Tool** — Flags, file processing, validation
- **Cloud-Native** — Kubernetes ConfigMaps, AWS, Vault, mTLS
- **Full Setup** — Schema generation and feature demonstration

Each example includes:
- Complete struct definitions
- Multiple configuration sources
- Comprehensive test suite (`examples_test.go`)
- Example configuration files (YAML, TOML)
- README with setup instructions

Run tests to see them in action:
```bash
go test ./examples -v
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

Built-in sources — pass any combination, first one to provide a value wins per field:

```go
confkit.FromYAML(path string) Source
confkit.FromYAMLOptional(path string) Source
confkit.FromYAMLFiles(paths ...string) Source
confkit.FromJSON(path string) Source
confkit.FromJSONFiles(paths ...string) Source
confkit.FromTOML(path string) Source
confkit.FromTOMLFiles(paths ...string) Source
confkit.FromEnv() Source
confkit.FromFlags() Source
confkit.FromFlagsWithArgs(args []string) Source

```

Optional sources (separate `go get` per module):

```go
// go get confkit/k8s
k8s.FromKubernetesConfigMap(namespace, name string) confkit.Source
k8s.FromKubernetesConfigMapWithPath(namespace, name, mountPath string) confkit.Source

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
aws.FromAWSSSMParameterStoreWithTTL(pathPrefix string, cacheTTL time.Duration) confkit.Source
aws.FromAWSSecretsManager(secretName string) confkit.Source
aws.FromAWSSecretsManagerWithRegion(secretName, region string) confkit.Source
aws.FromAWSSecretsManagerWithOptions(secretName, region string, cacheTTL time.Duration) confkit.Source
aws.FromAWSSecretsManagerMultiRegion(secretName string, regions []string) confkit.Source
aws.FromAWSSSMParameterStoreMultiRegion(pathPrefix string, regions []string) confkit.Source
```

---

## Validation

Rules used in the `validate` struct tag:

| Rule          | Types      | Behaviour                                           |
|---------------|------------|-----------------------------------------------------|
| `required`    | any        | Non-zero value required                             |
| `min=N`       | int, float | Value ≥ N                                           |
| `min=N`       | string     | Length ≥ N characters                               |
| `max=N`       | int, float | Value ≤ N                                           |
| `max=N`       | string     | Length ≤ N characters                               |
| `oneof=a b c` | string     | Value must equal one of the space-separated options |

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
func Load[T any](sources ...Source) (*T, error)

// Load with fine-grained options (validators, middleware, interpolation depth)
func LoadWithOptions[T any](options ...Option) (*T, error)
func LoadContext[T any](ctx context.Context, sources ...Source) (*T, error)
func LoadWithOptionsContext[T any](ctx context.Context, options ...Option) (*T, error)
func ValidateOnly[T any](ctx context.Context, options ...Option) (*T, error)
func MustLoad[T any](sources ...Source) *T
func MustLoadContext[T any](ctx context.Context, sources ...Source) *T

// Load and set up a file watcher in one call
func LoadWithWatcher[T any](filePath string, sources ...Source) (*T, *ConfigWatcher, error)

// Option constructors
func WithSource(source Source) Option
func WithValidator(name string, fn func (reflect.Value) error) Option
func WithModelValidator[T any](fn func (*T) error) Option
func WithMiddleware(fn MiddlewareFunc) Option
func WithAuditLogger(fn AuditLogger) Option
func WithLoadHook(fn LoadHookFunc) Option
func WithContext(ctx context.Context) Option
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

func (r *ErrorReport) Unwrap() []error

type FieldError struct {
    Path    string    // "Database.Password"
    Source  string    // "env", "yaml", "validation"
    Kind    ErrorKind // parse | validation | io
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

- Error messages show `<redacted>` for secret fields
- Validation values are redacted as `***REDACTED***`
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

### Prefix Mapping Rules

The `prefix` tag applies to all fields in a nested struct. When combined with `env` tags, the final environment variable name is constructed as: **`<prefix><env_tag_name>`**

| Struct Definition | env Tag | Prefix | Result Env Var |
|---|---|---|---|
| `Host string` | `env:"HOST"` | `prefix:"DB_"` | `DB_HOST` |
| `Port int` | `env:"PORT"` | `prefix:"DB_"` | `DB_PORT` |
| `timeout int` | `env:"TIMEOUT"` | `prefix:"CACHE_"` | `CACHE_TIMEOUT` |

If you omit the `env` tag on a field in a prefixed struct, the field **will not be loaded from environment variables** — the prefix alone is not enough. You must explicitly define `env:"FIELD_NAME"`.

```go
type Config struct {
    Database struct {
        Host     string `env:"HOST" default:"localhost"`   // reads DB_HOST
        Port     int    `env:"PORT" default:"5432"`        // reads DB_PORT
        User     string // ❌ not read from env (no tag)
        Password string `env:"PASSWORD" secret:"true"`     // reads DB_PASSWORD
    } `prefix:"DB_"`
}

// Valid env vars: DB_HOST, DB_PORT, DB_PASSWORD
// User must be set via config file or default
```

For multiple levels of nesting, prefixes accumulate:

```go
type Config struct {
    Main struct {
        Sub struct {
            Value string `env:"VALUE"` // reads: MAIN_SUB_VALUE
        } `prefix:"SUB_"`
    } `prefix:"MAIN_"`
}
```

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
    Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)
}

// FieldInfo fields available to your Lookup implementation:
// .Name         string            — "Password"
// .Path         string            — "Database.Password"
// .Type         reflect.Type
// .Value        reflect.Value
// .Tags         map[string]string — all struct tags
// .IsSecret     bool
// .HasDefault   bool
// .IsNested     bool
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


## Safe Dump API

```go
func Dump[T any](cfg T, opts ...DumpOption) ([]byte, error)
func DumpString[T any](cfg T, opts ...DumpOption) string
func DumpYAML[T any](cfg T, opts ...DumpOption) ([]byte, error)
```

---

## Supported Types

| Type                                              | Parsed from                       |
|---------------------------------------------------|-----------------------------------|
| `string`                                          | as-is                             |
| `int` / `int8` / `int16` / `int32` / `int64`      | decimal                           |
| `uint` / `uint8` / `uint16` / `uint32` / `uint64` | decimal                           |
| `float32` / `float64`                             | decimal                           |
| `bool`                                            | `true` `false` `1` `0` `yes` `no` |
| `time.Duration`                                   | `"5s"` `"1m30s"` `"2h"`           |
| `time.Time`                                       | RFC3339 `"2006-01-02T15:04:05Z07:00"` |
| `[]string`                                        | comma-separated `"a,b,c"`         |
| `[]int`                                           | comma-separated `"1,2,3"`         |
| `map[string]string` / `map[string]int` etc.       | `KEY=val,KEY2=val2` format        |

---

## Vault Auth Methods

```go
vault.VaultTokenAuth(token string) VaultAuth
vault.VaultAppRoleAuth(roleID, secretID string) VaultAuth
vault.VaultKubernetesAuth(role, jwt string) VaultAuth
```

---

## Documentation

- **[Full Documentation](https://confkit.xyz/)** — Getting started, guides, API reference
- **[Getting Started](https://confkit.xyz/docs/)** — 5-minute quick start
- **[API Reference](https://confkit.xyz/api/)** — Complete function and type reference
- **[Examples](https://confkit.xyz/examples/)** — Runnable code examples
- **[Sources Guide](https://confkit.xyz/docs/#sources)** — All configuration sources
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
