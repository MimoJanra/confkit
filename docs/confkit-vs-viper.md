# confkit vs Viper

**Use confkit** if you want struct-first, type-safe config loading with built-in validation and secret redaction—without the 50MB dependency bloat.

**Use Viper** if you need heavy runtime reloading with watches across many config files.

## Quick Comparison

|                          | confkit | Viper |
|--------------------------|:-------:|:-----:|
| **Typed `Load[T]`**       |    ✅    |   ❌   |
| **Defaults via tags**     |    ✅    |   ⚠️   |
| **Built-in validation**   |    ✅    |   ❌   |
| **Secret redaction**      |    ✅    |   ❌   |
| **Multi-source merging**  |    ✅    |   ✅   |
| **Lightweight core**      |    ✅    |   ❌   |
| **Cloud integrations**    |  optional | bundled |
| **Runtime reloading**     |    ✅    |   ✅   |

## Key Differences

### Type Safety
confkit returns your struct with full compile-time type information.
```go
cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))
// cfg is type Config, fully typed
```

Viper returns values via `GetString()`, `GetInt()` on a dynamic map.
```go
v := viper.New()
v.ReadInConfig()
port := v.GetInt("port")  // runtime lookup, potential panic
```

### Validation
confkit has built-in validation rules: `required`, `min`, `max`, `oneof`.
```go
type Config struct {
    Port int `env:"PORT" validate:"min=1,max=65535"`
}
```

Viper has no built-in validation—use a separate library.

### Secret Redaction
confkit automatically redacts secrets in errors and dumps.
```go
type Config struct {
    Token string `env:"API_TOKEN" secret:"true"`
}
// Errors show <redacted>, not the actual value
```

Viper has no built-in redaction—handle it manually.

### Dependencies
confkit: ~2 dependencies (yaml.v3, go-toml/v2)  
Viper: ~100+ dependencies bundled into core

## When to Choose

### Choose confkit if:
- You define config as a struct (not dynamic keys)
- You want validation without extra code
- You care about secret safety in error messages
- Your config is relatively stable (not constantly hot-reloading)
- You use cloud sources but want them optional

### Choose Viper if:
- You need to watch multiple files for changes
- Config keys are dynamic and unknown at compile time
- You want loose coupling between code and config
- You use spf13/cobra CLI framework

## Example: confkit

```go
type Config struct {
    Server struct {
        Host string `env:"HOST" default:"localhost"`
        Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    }
    Database struct {
        URL string `env:"DB_URL" validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))  // Human-readable error
}
```

## Example: Viper

```go
v := viper.New()
v.SetConfigFile("config.yaml")
v.ReadInConfig()
v.SetDefault("server.port", 8080)
// No type safety; need manual validation
host := v.GetString("server.host")
port := v.GetInt("server.port")
// Validation is your responsibility
if port < 1 || port > 65535 {
    log.Fatal("port out of range")
}
```

## Verdict

- **confkit wins** on: type safety, validation, secret redaction, lightweight core
- **Viper wins** on: dynamic config, file watching complexity, ecosystem maturity

confkit is production-ready v0.5.0 and solves the 80% case elegantly.
