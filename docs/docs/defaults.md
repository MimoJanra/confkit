---
layout: default
title: Defaults — confkit
---

# Defaults

Default values are specified in struct tags and applied only when no source provides a value for that field.

## Basic Defaults

Use the `default` struct tag:

```go
type Config struct {
    Host    string `env:"HOST" default:"localhost"`
    Port    int    `env:"PORT" default:"8080"`
    Timeout time.Duration `env:"TIMEOUT" default:"30s"`
    Debug   bool   `env:"DEBUG" default:"false"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
// If HOST not in env → uses "localhost"
// If PORT not in env → uses 8080
// If TIMEOUT not in env → uses 30 seconds
// If DEBUG not in env → uses false
```

## Type Coercion for Defaults

Defaults are specified as strings in tags but are coerced to the field type:

```go
type Config struct {
    Port     int           `default:"8080"`         // string "8080" → int 8080
    Timeout  time.Duration `default:"30s"`          // string "30s" → duration
    Enabled  bool          `default:"true"`         // string "true" → bool
    Numbers  []int         `default:"1,2,3"`        // string → []int
}
```

Supported type coercions:
- **int/uint/float**: decimal number strings
- **bool**: `true`, `false`, `yes`, `no`, `1`, `0`
- **time.Duration**: `"5s"`, `"1m30s"`, `"2h"`
- **string**: as-is
- **[]string**, **[]int**: comma-separated

## Default Behavior in Load Order

Defaults are applied **after all sources** are loaded. This means:

1. Load from all sources (YAML, env, flags, etc.)
2. Validate loaded values
3. Apply defaults **only** if field wasn't set by any source
4. Interpolate `${VAR}` references

```go
type Config struct {
    Host     string `env:"HOST" default:"localhost"`
    Port     int    `env:"PORT" default:"8080"`
}

cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                  // if env has PORT=9000 (higher priority)
    confkit.FromYAML("config.yaml"),    // if config.yaml has Host="example.com" (lower priority)
)
// Result: Host="example.com" (from YAML), Port=9000 (from env)
// Default "localhost" not used (YAML provided Host)
// Default 8080 not used (env provided PORT)
```

## Nested Struct Defaults

Defaults apply to all nested fields:

```go
type Config struct {
    Server struct {
        Host string `env:"HOST" default:"0.0.0.0"`
        Port int    `env:"PORT" default:"8080"`
    }
    Database struct {
        Host     string `env:"HOST" default:"localhost"`
        Port     int    `env:"PORT" default:"5432"`
        MaxConns int    `env:"MAX_CONNS" default:"10"`
    } `prefix:"DB_"`
}
```

When prefixed, env vars are `DB_HOST`, `DB_PORT`, `DB_MAX_CONNS`.

## String Interpolation with Defaults

Defaults can use `${VAR}` references to other fields or env vars:

```go
type Config struct {
    Host    string `env:"HOST" default:"localhost"`
    Port    int    `env:"PORT" default:"8080"`
    BaseURL string `env:"BASE_URL" default:"http://${HOST}:${PORT}"`
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
// If no env vars set:
//   Host → "localhost" (default)
//   Port → 8080 (default)
//   BaseURL → "http://localhost:8080" (interpolated)
```

Resolution order for `${VAR}`:
1. Other config fields
2. OS environment variables

See [String Interpolation](../docs/getting-started.md#string-interpolation) for details.

## Empty String vs No Default

If a field has no `default` tag and no source provides it, the field is zero-valued:

```go
type Config struct {
    Optional string  // no default
    Required string  `validate:"required"` // no default, but required!
}

cfg, err := confkit.Load[Config](confkit.FromEnv())
// Optional → "" (zero value)
// Required → error (validation fails)
```

To distinguish "not set" from "empty", use pointers or optionals.

## Defaults with Validation

Defaults must pass validation rules if both are specified:

```go
type Config struct {
    Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
}
```

If the default violates validation, confkit panics at startup (this is a programmer error).

```go
type Config struct {
    Port int `env:"PORT" default:"99999" validate:"min=1,max=65535"`
    // panic: default "99999" violates rule "max=65535"
}
```

**Best practice:** Ensure defaults always pass validation.

## Complex Type Defaults

### time.Duration

```go
type Config struct {
    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
    Wait     time.Duration `env:"WAIT" default:"1m30s"`
    Deadline time.Duration `env:"DEADLINE" default:"2h"`
}
```

Supported formats: `"5ms"`, `"30s"`, `"5m"`, `"1h"`, `"1h30m45s"`.

### Slices

```go
type Config struct {
    Ports    []int    `env:"PORTS" default:"8080,9090,9091"`
    Hosts    []string `env:"HOSTS" default:"localhost,127.0.0.1"`
}
```

Values are comma-separated. Whitespace is trimmed.

### Boolean

```go
type Config struct {
    Debug   bool `env:"DEBUG" default:"false"`
    Verbose bool `env:"VERBOSE" default:"true"`
}
```

Recognized: `true`, `false`, `yes`, `no`, `1`, `0`.

## Optional vs Required

Use defaults to make fields optional; use `validate:"required"` to enforce them:

```go
type Config struct {
    // Required (must be provided)
    DatabaseURL string `env:"DATABASE_URL" validate:"required"`

    // Optional with default
    Port int `env:"PORT" default:"8080"`

    // Optional with zero value (no default)
    LogFile string `env:"LOG_FILE"` // → ""
}
```

## Structured Defaults

If your struct has default values before calling `confkit.Load`, they **are not preserved** — confkit applies its own defaults:

```go
type Config struct {
    Port int `default:"8080"`
}

// This does NOT affect confkit:
cfg := Config{Port: 9999}

// confkit always uses its own defaults:
cfg, err := confkit.Load[Config](confkit.FromEnv())
// cfg.Port → 8080 (from tag), not 9999
```

To use zero values as "defaults", don't set a `default` tag:

```go
type Config struct {
    Port int // no default tag → 0 if not provided
}
```

## Default Metadata

Check which fields came from defaults:

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
fields := confkit.ScanFields(cfg)
for _, field := range fields {
    if field.HasDefault {
        fmt.Printf("%s has a default value\n", field.Name)
    }
}
```

## Environment Variable Precedence

Defaults are lowest priority — environment variables override them:

```
1. CLI flags     (highest)
2. Env vars
3. YAML/JSON file
4. Defaults      (lowest)
```

First source wins per field.

## Best Practices

1. **Provide sensible defaults**
   ```go
   Port int `env:"PORT" default:"8080"`
   ```

2. **Use validation with defaults**
   ```go
   Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
   ```

3. **Document required vs optional**
   ```go
   // Required
   DatabaseURL string `env:"DATABASE_URL" validate:"required"`
   // Optional
   LogLevel string `env:"LOG_LEVEL" default:"info"`
   ```

4. **Use interpolation for related defaults**
   ```go
   BaseURL string `env:"BASE_URL" default:"http://${HOST}:${PORT}"`
   ```

5. **Avoid magic defaults**
   ```go
   // ❌ Bad: Why 5? Unexplained.
   Retries int `env:"RETRIES" default:"5"`
   
   // ✅ Good: Clear default
   Retries int `env:"RETRIES" default:"3"` // sensible retry count
   ```

## Real-World Examples

See defaults in action across production examples:

- **[Web Service Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Complete configuration with sensible defaults
- **[Microservice Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Enterprise defaults for all services
- **[Cloud-Native Example](https://github.com/MimoJanra/confkit/tree/main/examples)** — Kubernetes resource defaults

All examples use a 3-level hierarchy: environment variables override config files, which override `defaults.yaml`.

## Next Steps

- **[Validation](./validation.md)** — Validation rules and custom validators
- **[Secret Redaction](./secret-redaction.md)** — Mark sensitive fields
- **[Examples](https://github.com/MimoJanra/confkit/tree/main/examples)** — Production-ready code with defaults
