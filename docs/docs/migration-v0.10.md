---
layout: default
title: Migration Guide - v0.10.0
---

# Migration Guide: v0.9 → v0.10

v0.10.0 introduces breaking changes. This guide explains what changed and how to migrate.

## Breaking Changes

### 1. Load[T] Now Returns Pointer (*T)

**Before (v0.9):**
```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
// cfg type: Config
```

**After (v0.10):**
```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
// cfg type: *Config (pointer)
```

**Why the change?**
- Go idiom for larger configs (often passed around, shared, or updated)
- More memory-efficient for large structures
- Aligns with patterns like `json.Unmarshal` returning `(*T, error)`

**Migration: Most code works unchanged**
Go automatically dereferences pointers for field access:
```go
// Both work — Go auto-dereferences
log.Printf("Port: %d", cfg.Port)   // ✅ Pointer dereference
log.Printf("Port: %d", (*cfg).Port) // ✅ Explicit dereference (unnecessary)
```

**Only change if you explicitly needed a value:**
```go
// Old code that relied on value type
func Handler(cfg Config) {}
handler(cfg)  // ERROR: type mismatch

// Fixed (v0.10+)
func Handler(cfg *Config) {}
handler(cfg)  // ✅ Works

// Or explicit dereference
handler(*cfg)  // ✅ Also works
```

### 2. LoadWithOptions[T] Returns Pointer

**Before:**
```go
cfg, err := confkit.LoadWithOptions[Config](options...)
// cfg type: Config
```

**After:**
```go
cfg, err := confkit.LoadWithOptions[Config](options...)
// cfg type: *Config (pointer)
```

Same migration as `Load[T]` above.

### 3. LoadWithWatcher[T] Returns Pointer

**Before:**
```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml", sources...)
// cfg type: Config
```

**After:**
```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml", sources...)
// cfg type: *Config (pointer)
```

Same migration as `Load[T]` above.

### 4. OTel Module Functions Return Pointers

If using `github.com/MimoJanra/confkit/otel`:

**Before:**
```go
cfg, err := otel.Load[Config](tracer, sources...)
// cfg type: Config
```

**After:**
```go
cfg, err := otel.Load[Config](tracer, sources...)
// cfg type: *Config (pointer)
```

## New Features in v0.10

### FromYAMLOptional()

Load YAML without failing if the file doesn't exist:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    confkit.FromYAMLOptional("config.local.yaml"),  // Optional
    confkit.FromYAML("config.yaml"),               // Required
)
```

### Automatic snake_case Mapping

Fields without explicit tags now auto-match snake_case versions in YAML/JSON/TOML:

```go
type Config struct {
    DatabaseURL string  // Auto-matches "database_url" in config
    ShutdownSecs int    // Auto-matches "shutdown_secs" in config
}

// config.yaml
database_url: postgres://localhost/db
shutdown_secs: 30
```

No more need for explicit tags unless you want a different name!

### Enhanced Error Sources

All errors now consistently include the `source` field:

```
Invalid configuration:

  DatabaseURL
    error: field is required
    source: validation
```

## Step-by-Step Migration

1. **Update go.mod:**
   ```bash
   go get -u github.com/MimoJanra/confkit@latest
   ```

2. **Recompile:**
   ```bash
   go build ./...
   ```
   Most code will compile without changes due to Go's auto-dereference behavior.

3. **Fix any compilation errors:**
   - Functions expecting `Config` value now need `*Config`
   - Or pass `*cfg` if the function signature can't change
   - Or receive pointer and dereference: `func Handler(cfg *Config) { ... }`

4. **Optional: Improve with new features**
   - Remove redundant `yaml:` tags using auto snake_case mapping
   - Use `FromYAMLOptional()` for optional local configs
   - Rely on new consistent error sources in error handling

## Compatibility Notes

- **Go version:** Requires Go 1.24.0+
- **Dependencies:** No new dependencies
- **Submodules:** All submodules (Vault, Consul, AWS, etc.) updated to match
- **Core API:** Only the return type changed; all methods remain the same

## Minimal Migration Example

```go
// v0.9
func main() {
    cfg, err := confkit.Load[Config](confkit.FromEnv())
    if err != nil {
        log.Fatal(err)
    }
    server := &Server{Config: cfg}  // Store in struct
}

// v0.10 — same code, no changes needed!
func main() {
    cfg, err := confkit.Load[Config](confkit.FromEnv())
    if err != nil {
        log.Fatal(err)
    }
    server := &Server{Config: cfg}  // cfg is *Config, but struct field can be *Config
}
```

## Frequently Asked Questions

**Q: Will my code break?**
A: Most code will work unchanged. Only explicit type matches (like function parameters expecting `Config` value) will need updates.

**Q: How do I get the value type?**
A: Dereference: `cfgValue := *cfg`

**Q: Why the breaking change?**
A: Aligns with Go idioms. Pointers are more idiomatic for configs that are larger or shared.

**Q: Is there a way to get the old behavior?**
A: No, but migration is simple. Most code works unchanged due to Go's auto-dereference behavior.

## Next Steps

- See **[FromYAMLOptional()](./sources.md#yaml-files)** for optional file loading
- See **[Automatic snake_case mapping](./sources.md#automatic-snake_case-mapping-v010)** for field mapping
- See **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** for production-ready code with v0.10
