
# confkit vs Viper

confkit is a type-safe configuration toolkit for Go projects that want struct-based config loading, defaults, validation, source merging, and secret redaction. Viper is a mature, flexible configuration library focused on runtime reloading and dynamic configuration.

## Key Differences

### Type Safety

**confkit:** Returns your struct with full type information checked at compile time.

```go
cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))
// cfg is exactly type Config, fully typed
```

**Viper:** Returns values via `GetString()`, `GetInt()`, etc. on a dynamic map.

```go
v := viper.New()
v.SetConfigFile("config.yaml")
v.ReadInConfig()
port := v.GetInt("port")  // string key, runtime lookup
```

### Built-in Defaults & Validation

**confkit:** Via struct tags — no extra code.

```go
type Config struct {
    Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
}
```

**Viper:** Manual setup required for each field.

```go
viper.SetDefault("port", 8080)
// Validation is not built-in; use a separate library
```

### Error Messages

**confkit:** Human-readable, field-focused errors.

```
Invalid configuration:

  Port
    error: must be between 1 and 65535, got 99999
    source: env (PORT)
```

**Viper:** Minimal error context — you build custom handlers.

### Secret Redaction

**confkit:** Mark fields as `secret:"true"`, they're automatically redacted in errors and dumps.

```go
type Config struct {
    Token string `env:"API_TOKEN" secret:"true"`
}
// Errors and logs automatically show <redacted> for this field
```

**Viper:** No built-in secret redaction — custom handling required.

### Multi-Source Merging

**confkit:** Explicit source precedence — first source wins per field.

```go
cfg, err := confkit.Load[Config](
    confkit.FromFlags(),                  // highest
    confkit.FromEnv(),                    // overrides file
    confkit.FromYAML("defaults.yaml"),    // lowest
)
```

**Viper:** Also supports multi-source, but requires method calls and less explicit precedence.

```go
viper.AddConfigPath(".")
viper.SetConfigName("config")
viper.ReadInConfig()
// Precedence is less clear; binding and watching can interact unpredictably
```

### Cloud Sources

**confkit:** Optional modules only when needed (Vault, Consul, etcd, AWS).

```bash
go get github.com/MimoJanra/confkit/vault@latest
```

**Viper:** Cloud integrations are bundled or require external libraries (increases binary size).

### Runtime Reloading

**confkit:** File watching with `LoadWithWatcher()`.

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml", sources...)
watcher.AddListener(func(old, new any, err error) { ... })
watcher.Start()
```

**Viper:** Heavy runtime watching with `viper.OnConfigChange()`.

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) { ... })
```

Viper excels at watching multiple files and hot-reloading across a complex hierarchy. confkit keeps it simpler.

## When to Use confkit

- You want a single struct definition for your entire config
- You need defaults and validation without extra code
- You care about safe error messages (no secret leaks in logs)
- Your config is relatively stable (not constantly changing at runtime)
- You use cloud sources but want them optional, not bundled
- You prefer explicit source precedence

## When to Use Viper

- You need heavy runtime reloading with watches across many files
- Your application is highly dynamic and config changes frequently
- You want to avoid struct-based config entirely
- You need loose coupling between code and config keys
- You need existing integrations (e.g., with spf13/cobra CLI framework)

## Feature Comparison Table

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
| **Dynamic keys**          |    ❌    |   ✅   |
| **Schema generation**     |    ✅    |   ❌   |

## Migration Path

If you're moving from Viper to confkit:

1. Define your config as a struct
2. Add struct tags (`env`, `flag`, `default`, `validate`)
3. Replace `viper.ReadInConfig()` with `confkit.Load[T]()`
4. Replace `viper.GetString()` calls with direct struct field access
5. For runtime reloading, use `LoadWithWatcher()`

## Maturity

- **Viper:** Mature, widely used, large community
- **confkit:** Production-ready v0.9.0, focused on a specific use case (type-safe, validated, struct-first)
