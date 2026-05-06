# confkit vs envconfig

**Use confkit** if you want struct-first, type-safe config loading with validation, defaults, and secret redaction—from multiple sources.

**Use envconfig** if you only care about environment variables and want the smallest possible library.

## Quick Comparison

|                          | confkit | envconfig |
|--------------------------|:-------:|:---------:|
| **Typed struct loading**  |    ✅    |     ✅     |
| **Environment variables** |    ✅    |     ✅     |
| **YAML/JSON/TOML files**  |    ✅    |     ❌     |
| **CLI flags**             |    ✅    |     ❌     |
| **Defaults via tags**     |    ✅    |     ❌     |
| **Validation rules**      |    ✅    |     ❌     |
| **Secret redaction**      |    ✅    |     ❌     |
| **Cloud sources**         |  optional |    ❌     |
| **Error context**         |    ✅    |     ⚠️    |
| **Multi-source merging**  |    ✅    |     ❌     |

## Key Differences

### Scope
**confkit:** Full configuration toolkit—multiple sources, validation, defaults, redaction.

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    confkit.FromFlags(),
)
```

**envconfig:** Environment variables only. No files, no validation built-in, no defaults.

```go
var cfg Config
envconfig.Process("APP", &cfg)
// Only reads from os.Getenv()
```

### Defaults
**confkit:**
```go
type Config struct {
    Port int `env:"PORT" default:"8080"`
}
```

**envconfig:** No `default` tag support—set zero values or post-process.

### Validation
**confkit:** Built-in rules: `required`, `min`, `max`, `oneof`.
```go
type Config struct {
    Port int `env:"PORT" validate:"min=1,max=65535"`
}
```

**envconfig:** No validation—write custom code after parsing.

### Error Messages
**confkit:** Structured, human-readable.
```
Invalid configuration:
  Port
    error: must be between 1 and 65535
    source: env (PORT)
```

**envconfig:** Basic Go errors, less context.

### Secret Redaction
**confkit:** Automatic via `secret:"true"` tag.

**envconfig:** No redaction—passwords leak in error messages.

## When to Choose

### Choose confkit if:
- You need multiple sources (YAML, env, flags, cloud)
- You want defaults and validation built-in
- You care about error messages and debugging
- You use cloud sources (Vault, AWS, Kubernetes)
- You need secret redaction

### Choose envconfig if:
- You only load config from environment variables
- You want the smallest library possible
- You're OK writing validation code yourself
- You don't need cloud sources
- Your config is simple (no defaults, no validation)

## Example: confkit

```go
type Config struct {
    Port     int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    Database string `env:"DATABASE_URL" validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

## Example: envconfig

```go
type Config struct {
    Port     int    `envconfig:"PORT"`
    Database string `envconfig:"DATABASE_URL"`
}

var cfg Config
envconfig.Process("", &cfg)
// Must manually set defaults
if cfg.Port == 0 {
    cfg.Port = 8080
}
// Must manually validate
if cfg.Port < 1 || cfg.Port > 65535 {
    log.Fatal("port out of range")
}
```

## Verdict

- **confkit wins** on: feature completeness, validation, defaults, error context, security
- **envconfig wins** on: minimalism, stdlib-only dependencies

confkit is production-ready v0.5.0 and handles the full config lifecycle.
