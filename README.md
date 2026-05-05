# confkit

Typed configuration loading for Go. Define your config as a struct, declare sources, get a fully validated value back — or a clear error message telling you exactly what's wrong and where it came from.

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

**Error output when something is wrong:**
```
Invalid configuration:

  DSN
    source: env
    error: field is required
```

---

## Install

```bash
go get confkit
```

> Module name will be updated to a public path when the package is published.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Sources](#sources)
- [Struct Tags](#struct-tags)
- [Source Precedence](#source-precedence)
- [Validation](#validation)
- [Secrets](#secrets)
- [String Interpolation](#string-interpolation)
- [Nested Structs](#nested-structs)
- [Hot Reload](#hot-reload)
- [Enterprise Sources](#enterprise-sources)
- [Custom Sources](#custom-sources)
- [Custom Validators](#custom-validators)
- [Schema Generation](#schema-generation)
- [Supported Types](#supported-types)

---

## Quick Start

```go
package main

import (
    "log"
    "confkit"
)

type Config struct {
    Host string `env:"HOST" default:"localhost"`
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DB   struct {
        DSN      string `env:"DB_DSN" validate:"required" secret:"true"`
        MaxConns int    `env:"DB_MAX_CONNS" default:"10" validate:"min=1,max=100"`
    }
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromYAML("config.yaml"),
        confkit.FromEnv(),
        confkit.FromFlags(),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }

    log.Printf("Listening on %s:%d", cfg.Host, cfg.Port)
}
```

---

## Sources

Sources are queried in the order you pass them. The **last source wins** for each field.

| Function | Description |
|---|---|
| `FromYAML(path)` | YAML file |
| `FromJSON(path)` | JSON file |
| `FromTOML(path)` | TOML file |
| `FromEnv()` | OS environment variables |
| `FromFlags()` | CLI flags (`--flag-name value`) |
| `FromKubernetesConfigMap(namespace, name)` | Kubernetes ConfigMap (mounted as files) |

```go
// Typical layered setup: file → env overrides → CLI overrides
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    confkit.FromFlags(),
)
```

---

## Struct Tags

| Tag | Description | Example |
|---|---|---|
| `env` | Environment variable name | `env:"DATABASE_URL"` |
| `flag` | CLI flag name | `flag:"port"` |
| `short` | Short CLI flag (single character) | `short:"p"` |
| `default` | Default value if no source provides one | `default:"8080"` |
| `validate` | Validation rules | `validate:"required,min=1"` |
| `secret` | Redact value in errors and dumps | `secret:"true"` |
| `prefix` | Env variable prefix for a nested struct | `prefix:"DB_"` |
| `help` | Description for schema/CLI help output | `help:"Listen port"` |
| `hidden` | Hide from CLI help | `hidden:"true"` |

---

## Source Precedence

Sources are applied left to right — later sources override earlier ones:

```go
confkit.Load[Config](
    confkit.FromYAML("config.yaml"), // lowest priority
    confkit.FromEnv(),
    confkit.FromFlags(),              // highest priority
)
```

If a field is not found in any source and has no `default`, it is left at the zero value. Add `validate:"required"` to make it an error.

---

## Validation

Built-in rules in the `validate` tag:

| Rule | Applies to | Description |
|---|---|---|
| `required` | any | Field must not be zero/empty |
| `min=N` | int, float, string | Minimum value or minimum string length |
| `max=N` | int, float, string | Maximum value or maximum string length |
| `oneof=a b c` | string | Value must be one of the listed options |

```go
type Config struct {
    Port    int    `env:"PORT"    validate:"required,min=1,max=65535"`
    Level   string `env:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
    Timeout int    `env:"TIMEOUT" validate:"min=1,max=300"`
}
```

### Custom Validators

Use `LoadWithOptions` to register a validator for the duration of a single load:

```go
import "reflect"

cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("positive", func(v reflect.Value) error {
        if v.Kind() == reflect.Int && v.Int() <= 0 {
            return fmt.Errorf("must be positive")
        }
        return nil
    }),
)
```

Reference it in a tag:

```go
type Config struct {
    Workers int `env:"WORKERS" validate:"positive"`
}
```

---

## Secrets

Mark a field with `secret:"true"` and confkit will redact its value everywhere — in errors, in `DumpConfig`, and in logs.

```go
type Config struct {
    APIKey   string `env:"API_KEY"   secret:"true" validate:"required"`
    Password string `env:"DB_PASS"   secret:"true"`
}
```

If `APIKey` is missing:
```
Invalid configuration:

  APIKey
    source: env
    error: field is required
```

If `Password` fails to parse, the bad value is shown as `<redacted>` instead of the actual string.

### Config Dump

```go
fields := confkit.ScanFields(cfg)
data, err := confkit.DumpConfig(cfg, fields)
// {"APIKey": "***REDACTED***", "Host": "localhost", ...}
```

---

## String Interpolation

Reference environment variables or other config fields using `${VAR}` syntax:

```go
type Config struct {
    Host    string `env:"HOST"    default:"localhost"`
    Port    int    `env:"PORT"    default:"8080"`
    BaseURL string `env:"BASE_URL" default:"http://${HOST}:${PORT}"`
}
```

`BaseURL` resolves to `http://localhost:8080` if `HOST` and `PORT` are not overridden.

References are resolved from:
1. Other config field values
2. OS environment variables

Circular references are detected and reported as errors.

---

## Nested Structs

confkit scans nested structs recursively. Use `prefix` on a nested struct to namespace its environment variables:

```go
type Config struct {
    App      AppConfig
    Database DatabaseConfig `prefix:"DB_"`
    Cache    CacheConfig    `prefix:"CACHE_"`
}

type DatabaseConfig struct {
    Host     string `env:"HOST" default:"localhost"`
    Port     int    `env:"PORT" default:"5432"`
    Password string `env:"PASSWORD" secret:"true" validate:"required"`
}
```

With `prefix:"DB_"`, the `Host` field is read from `DB_HOST`, `Port` from `DB_PORT`, etc.

---

## Hot Reload

Watch a config file for changes without restarting the process:

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
        log.Printf("reload error: %v", err)
        return
    }
    log.Println("config reloaded")
})

watcher.SetPollInterval(5 * time.Second)
watcher.Start()
defer watcher.Stop()
```

The watcher polls the file for modification time changes. When a change is detected, all registered listeners are called with the old and new config values.

---

## Enterprise Sources

Enterprise sources live in separate optional modules — import only what you need. Each module requires its own SDK, which is why they are separated from the core.

### HashiCorp Vault

```bash
go get confkit/vault
```

```go
import "confkit/vault"

cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    vault.FromVault("https://vault.example.com", vault.VaultTokenAuth(os.Getenv("VAULT_TOKEN")), "myapp"),
)
```

Auth methods: `VaultTokenAuth(token)`, `VaultAppRoleAuth(roleID, secretID)`, `VaultKubernetesAuth(role, jwt)`.

KV v1 vs v2:
```go
vault.FromVaultWithKVVersion(addr, auth, 1, "myapp") // KV v1
vault.FromVaultWithKVVersion(addr, auth, 2, "myapp") // KV v2 (default)
```

### HashiCorp Consul

```bash
go get confkit/consul
```

```go
import "confkit/consul"

cfg, err := confkit.Load[Config](
    consul.FromConsul("localhost:8500"),
    // or with ACL token:
    consul.FromConsulWithToken("localhost:8500", os.Getenv("CONSUL_TOKEN")),
    // or with datacenter:
    consul.FromConsulWithOptions("localhost:8500", token, "dc1"),
)
```

### etcd

```bash
go get confkit/etcd
```

```go
import "confkit/etcd"

cfg, err := confkit.Load[Config](
    etcd.FromEtcd([]string{"localhost:2379"}),
    // or with a key prefix:
    etcd.FromEtcdWithPrefix([]string{"localhost:2379"}, "/myapp/"),
)
```

### AWS

```bash
go get confkit/aws
```

**SSM Parameter Store:**
```go
import "confkit/aws"

cfg, err := confkit.Load[Config](
    aws.FromAWSSSMParameterStore("/myapp/"),
)
```

**Secrets Manager:**
```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManager("myapp/production"),
    // or with explicit region:
    aws.FromAWSSecretsManagerWithRegion("myapp/production", "us-east-1"),
)
```

**Multi-region failover:**
```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManagerMultiRegion("myapp/production", []string{
        "us-east-1",
        "us-west-2",
        "eu-west-1",
    }),
)
```

---

## Custom Sources

Implement the `Source` interface to plug in any backend:

```go
type Source interface {
    Name() string
    Lookup(field *confkit.FieldInfo) (any, bool, error)
}
```

```go
type RedisSource struct {
    client *redis.Client
    prefix string
}

func (r *RedisSource) Name() string { return "redis" }

func (r *RedisSource) Lookup(field *confkit.FieldInfo) (any, bool, error) {
    key := r.prefix + strings.ToLower(field.Path)
    val, err := r.client.Get(context.Background(), key).Result()
    if errors.Is(err, redis.Nil) {
        return "", false, nil
    }
    if err != nil {
        return "", false, err
    }
    return val, true, nil
}

// Use it like any built-in source:
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    &RedisSource{client: rdb, prefix: "config:"},
)
```

`FieldInfo` gives you everything you need:

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Field name (`"Port"`) |
| `Path` | `string` | Dot-separated path (`"Database.Port"`) |
| `Tags` | `map[string]string` | All struct tags for this field |
| `IsSecret` | `bool` | True if `secret:"true"` |
| `HasDefault` | `bool` | True if `default:"..."` is set |

---

## Schema Generation

Generate a JSON Schema or Markdown documentation from your config struct:

```go
import "confkit/schema"

s, err := schema.GenerateSchema[Config]()
if err != nil {
    log.Fatal(err)
}

// JSON Schema
data, _ := json.MarshalIndent(s, "", "  ")
fmt.Println(string(data))

// Markdown table
fmt.Println(schema.GenerateMarkdown[Config]())

// CLI help text
fmt.Println(schema.GenerateCLIHelp[Config]())
```

---

## Supported Types

| Go type | Parsed from string |
|---|---|
| `string` | as-is |
| `int`, `int8` … `int64` | decimal integer |
| `uint`, `uint8` … `uint64` | unsigned integer |
| `float32`, `float64` | floating-point |
| `bool` | `true`, `false`, `1`, `0`, `yes`, `no` |
| `time.Duration` | `"5s"`, `"1m30s"`, `"2h"` |
| `[]string` | comma-separated: `"a,b,c"` |
| `[]int` | comma-separated: `"1,2,3"` |

---

## License

MIT
