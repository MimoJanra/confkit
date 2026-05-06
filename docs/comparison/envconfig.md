
# confkit vs envconfig

confkit is a type-safe configuration toolkit for Go projects that want struct-based config loading, defaults, validation, source merging, and secret redaction. envconfig is a lightweight library that parses struct tags to load configuration from environment variables.

## Key Differences

### Scope

**confkit:** Full configuration toolkit — multiple sources, validation, defaults, secret redaction, all built-in.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    confkit.FromFlags(),
)
```

**envconfig:** Environment variables only — no files, no validation, no defaults in the library.

```go
var cfg Config
err := envconfig.Process("APP", &cfg)
// Only reads from os.Getenv(), no other sources
```

### Defaults

**confkit:** Via struct tags.

```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}
```

**envconfig:** No built-in defaults — you set zero values or post-process.

```go
type Config struct {
    Port int `envconfig:"PORT"`  // No default support
}
// Must manually set defaults after parsing
```

### Validation

**confkit:** Built-in validation rules.

```go
type Config struct {
    Port int `env:"PORT" validate:"min=1,max=65535"`
}
```

**envconfig:** No built-in validation — you write custom code.

```go
var cfg Config
envconfig.Process("APP", &cfg)
if cfg.Port < 1 || cfg.Port > 65535 {
    return fmt.Errorf("port out of range")
}
```

### Error Messages

**confkit:** Structured, human-readable errors with context.

```
Invalid configuration:

  Database.Port
    error: must be between 1 and 65535
    source: env (DATABASE_PORT)
```

**envconfig:** Basic Go errors, less context.

```
port: parsing "invalid": invalid syntax
```

### Secret Redaction

**confkit:** Automatic via `secret:"true"` tag.

```go
type Config struct {
    Password string `env:"DB_PASSWORD" secret:"true"`
}
// Errors and dumps automatically redact this field
```

**envconfig:** No secret redaction — passwords appear in error messages.

### Multi-Source Loading

**confkit:** Mix YAML, env, flags, Kubernetes, AWS, Vault, etc. with explicit precedence.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("defaults.yaml"),
    confkit.FromEnv(),
)
```

**envconfig:** Environment variables only. To combine sources, you write custom code.

```go
// No built-in support; load each source separately and merge manually
```

### Struct Prefixes

**confkit:** Via struct tags — any nesting level.

```go
type Config struct {
    DB struct {
        Host string `env:"HOST"`
    } `prefix:"DB_"`
}
// Reads DB_HOST from env
```

**envconfig:** Also supports prefixes, similar approach.

```go
type Config struct {
    DB struct {
        Host string `envconfig:"HOST"`
    } `envconfig:"DB"`
}
```

### Type Support

**confkit:** Primitives, slices, time.Duration, nested structs, custom parsers.

**envconfig:** Primitives, slices, basic types. Similar but narrower.

## When to Use confkit

- You need multiple sources (YAML, env, flags, cloud)
- You want defaults and validation built-in
- You care about error messages and debugging
- You use cloud sources (Vault, AWS, Kubernetes)
- You need secret redaction
- You want a single struct to describe your entire config

## When to Use envconfig

- You only care about environment variables
- You want the absolute smallest library for env parsing
- Your config is simple (no defaults, no validation)
- You're OK writing validation code yourself
- You don't need cloud sources

## Feature Comparison Table

|                          | confkit | envconfig |
|--------------------------|:-------:|:---------:|
| **Typed struct loading**  |    ✅    |     ✅     |
| **Environment variables**  |    ✅    |     ✅     |
| **YAML/JSON/TOML files**   |    ✅    |     ❌     |
| **CLI flags**             |    ✅    |     ❌     |
| **Defaults via tags**     |    ✅    |     ❌     |
| **Built-in validation**   |    ✅    |     ❌     |
| **Secret redaction**      |    ✅    |     ❌     |
| **Cloud sources**         |  optional |    ❌     |
| **Error context**         |    ✅    |     ⚠️    |
| **Struct prefixes**       |    ✅    |     ✅     |
| **Multi-source merging**  |    ✅    |     ❌     |
| **Lightweight**           |    ✅    |     ✅     |

## Migration Path

If you're moving from envconfig to confkit:

1. Keep your struct definitions as-is
2. Add support for other sources:
   ```go
   confkit.Load[Config](
       confkit.FromEnv(),
       // Add YAML, flags, etc.
   )
   ```
3. Add `default` tags:
   ```go
   type Config struct {
       Port int `env:"PORT" default:"8080"`
   }
   ```
4. Add `validate` tags:
   ```go
   type Config struct {
       Port int `env:"PORT" validate:"min=1,max=65535"`
   }
   ```
5. Replace `envconfig.Process()` with `confkit.Load[T]()`

## Library Size & Dependencies

| Library    | Dependencies | Binary overhead |
|------------|:------------:|:---------------:|
| confkit    | yaml.v3, go-toml/v2 | ~500KB (core) |
| envconfig  | stdlib only  | ~100KB |

confkit adds dependencies for YAML and TOML support; envconfig is pure stdlib.

For cloud sources, confkit lets you opt-in. envconfig has no cloud support.

## Use Case Example

**Simple CLI tool:** envconfig is perfect.

```go
type Config struct {
    ApiKey string `envconfig:"API_KEY"`
}
```

**Production service with config files, env overrides, validation, and Vault:** confkit.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    vault.FromVault(...),  // optional
)
```

## Maturity

- **envconfig:** Lightweight, stable, minimal changes
- **confkit:** Production-ready v0.5.0, actively developed with cloud integrations
