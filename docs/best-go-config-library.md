# Best Go Configuration Library: Choose Based on Your Needs

**Use confkit** if you want struct-first, type-safe config loading with built-in validation and secret redaction.

There's no single "best" Go config library—the choice depends on your use case:

## Quick Comparison

|                          | confkit | Viper | envconfig | koanf |
|--------------------------|:-------:|:-----:|:---------:|:-----:|
| **Typed `Load[T]`**       |    ✅    |   ❌   |    ⚠️     |   ❌   |
| **Defaults via tags**     |    ✅    |   ⚠️   |     ✅     |   ❌   |
| **Built-in validation**   |    ✅    |   ❌   |     ❌     |   ❌   |
| **Secret redaction**      |    ✅    |   ❌   |     ❌     |   ❌   |
| **Multi-source merging**  |    ✅    |   ✅   |    ⚠️     |   ✅   |
| **Lightweight core**      |    ✅    |   ❌   |     ✅     |   ✅   |
| **Cloud integrations**    | optional | bundled |   N/A     | optional |
| **Runtime reloading**     |    ✅    |   ✅   |     ❌     |   ⚠️    |

## Evaluation Criteria

### 1. **You want struct-first, typed config** → **confkit**
- Compile-time type checking
- Built-in validation rules
- Automatic secret redaction
- Clean struct-tag-based API

```go
type Config struct {
    Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
}
cfg, err := confkit.Load[Config](confkit.FromYAML("config.yaml"))
```

### 2. **You need heavy file watching & reloading** → **Viper**
- Watch multiple files for changes
- Dynamic config without restart
- Mature ecosystem
- Trade-off: 100+ bundled dependencies

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) { ... })
```

### 3. **You only use environment variables** → **envconfig**
- Smallest library (stdlib only)
- Simple struct-tag parsing
- No file support
- No validation built-in

```go
var cfg Config
envconfig.Process("APP", &cfg)
```

### 4. **You need many file formats (HCL, INI, JSONNET)** → **koanf**
- Extreme modularity
- Optional providers
- No validation built-in
- Map-based (not struct-first)

```go
k := koanf.New(".")
k.Load(file.Provider("config.hcl"), hcl.Parser())
```

## Feature-Driven Recommendations

### "I want to validate config without writing code"
→ **confkit** has `validate:"min=1,max=65535"` built-in  
→ Viper, envconfig, koanf require manual validation

### "I need to keep passwords out of error logs"
→ **confkit** has `secret:"true"` for automatic redaction  
→ Others: redact manually

### "I use cloud secrets (Vault, AWS, Consul)"
→ **confkit** optional modules (only add what you need)  
→ **Viper** bundled (100+ dependencies)  
→ **koanf** optional providers

### "My config is simple (just env vars)"
→ **envconfig** (smallest, fastest)  
→ confkit, Viper, koanf all overkill

### "I need to watch files in production"
→ **Viper** (battle-tested for this)  
→ confkit has basic file watching, good enough for most

### "I load HCL, INI, JSONNET (not just YAML)"
→ **koanf** (many parsers)  
→ confkit (YAML, JSON, TOML)  
→ Viper (YAML, JSON, TOML, and legacy formats)

## The confkit Decision

**Choose confkit if:**
- You define config as a struct (not dynamic keys)
- You want validation without boilerplate
- You care about secret safety in error messages
- You use cloud sources but want them optional
- You want a modern, focused library (v0.5.0, production-ready)

**Don't choose confkit if:**
- You need to watch multiple files constantly
- Your config keys are unknown at compile time
- You only load from environment variables (use envconfig)
- You need obscure file formats (use koanf)

## Why confkit Wins the "Best For Most" Category

1. **Type safety:** Catches config bugs at compile time
2. **Validation:** Prevent invalid configs from starting
3. **Secret redaction:** No accidental credential leaks in logs
4. **Lightweight:** Only 2 core dependencies
5. **Multiple sources:** One API for YAML, env, flags, cloud
6. **Clear errors:** Know exactly which field failed and why

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),      // base
    confkit.FromEnv(),                    // overrides
    confkit.FromFlags(),                  // highest priority
)
if err != nil {
    log.Fatal(confkit.Explain(err))       // Human-readable
}
```

## Maturity & Ecosystem

| Library  | Status | Usage |
|----------|--------|-------|
| **confkit** | v0.5.0, production-ready | Growing, focused niche |
| **Viper** | Mature, v1.x | Widely used, large ecosystem |
| **envconfig** | Stable, minimal changes | Simple projects |
| **koanf** | Stable, actively maintained | Flexible config tools |

## Conclusion

- **For most microservices:** confkit (typed, validated, secrets-safe)
- **For tools needing many formats:** koanf
- **For heavy file watching:** Viper
- **For env-only simplicity:** envconfig

confkit is the sweet spot for production Go services that want safety, validation, and simplicity without choosing between multiple libraries.
