# confkit vs koanf

**Use confkit** if you want struct-first, type-safe config loading with built-in validation and secret redaction.

**Use koanf** if you need extreme modularity, many file formats (HCL, INI, JSONNET), and don't need validation.

## Quick Comparison

|                          | confkit | koanf |
|--------------------------|:-------:|:-----:|
| **Typed `Load[T]`**       |    ✅    |   ❌   |
| **Defaults via tags**     |    ✅    |   ❌   |
| **Built-in validation**   |    ✅    |   ❌   |
| **Secret redaction**      |    ✅    |   ❌   |
| **Multi-source merging**  |    ✅    |   ✅   |
| **Map-based access**      |    ❌    |   ✅   |
| **Many file formats**     |    ⚠️    |   ✅   |
| **Custom providers**      |  limited |   ✅   |
| **Schema generation**     |    ✅    |   ❌   |
| **Error context**         |    ✅    |   ⚠️   |
| **Cloud integrations**    |  optional | optional |

## Key Differences

### Architecture
**confkit:** Struct-first—your config is a Go struct, everything flows from there.

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL" secret:"true"`
}

cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))
```

**koanf:** Map-first—your config is a map, you unmarshal to struct afterward.

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
var cfg Config
k.Unmarshal("", &cfg)
```

### Type Safety
**confkit:** Full compile-time type checking—no runtime casts.

```go
cfg.Port  // type int, no casting
```

**koanf:** Loose initially, cast at runtime.

```go
port := k.Get("port").(int)  // Can panic
```

### Validation
**confkit:** Built-in rules—`required`, `min`, `max`, `oneof`.

```go
Port int `validate:"min=1,max=65535"`
```

**koanf:** No validation—you handle it after unmarshaling.

### Error Messages
**confkit:** Structured, human-readable.

**koanf:** Generic unmarshaling errors from encoding/json.

### Secret Redaction
**confkit:** Automatic via `secret:"true"` tag.

**koanf:** No built-in redaction—redact manually.

### File Formats
**confkit:** YAML, JSON, TOML, env, flags.

**koanf:** YAML, JSON, TOML, HCL, INI, JSONNET, and custom providers.

## When to Choose

### Choose confkit if:
- You define config as a struct (known at compile time)
- You need validation and defaults built-in
- You care about type safety and error messages
- You use cloud sources (Vault, AWS, Kubernetes)
- You need secret redaction
- Your config structure is stable and well-defined

### Choose koanf if:
- You need support for many file formats (HCL, INI, JSONNET, etc.)
- Your config is dynamic (fields come and go)
- You prefer map-based access with loose typing
- You want extreme modularity and custom providers
- You're building a tool that accepts arbitrary config formats

## Example: confkit

```go
type Config struct {
    Server struct {
        Host string `env:"HOST" default:"localhost"`
        Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    }
    Database struct {
        URL string `env:"URL" validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

## Example: koanf

```go
k := koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Load(env.Provider("APP_", ".", nil), nil)

var cfg Config
k.Unmarshal("", &cfg)

// Manual validation
if cfg.Port < 1 || cfg.Port > 65535 {
    log.Fatal("port out of range")
}
```

## Verdict

- **confkit wins** on: type safety, validation, defaults, secret redaction, error messages
- **koanf wins** on: file format variety, modularity, dynamic config access

confkit is production-ready v0.5.0 and ideal for typed, validated microservices.
