---
layout: default
title: Validation — confkit
---

# Validation

confkit includes a built-in validation engine. Validation rules are defined in struct tags and applied after all sources are loaded but before defaults are applied.

## Built-in Validators

### `required`

Field must have a non-zero value.

```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" validate:"required"`
    // Error if DATABASE_URL is empty or not set
}
```

Works on all types:
- `string`: must be non-empty
- `int`, `float`: must be non-zero
- `bool`: must be true
- slices: must have at least one element

### `min=N`

Minimum value (for numbers) or minimum length (for strings).

```go
type Config struct {
    Port     int    `env:"PORT"  validate:"min=1,max=65535"`
    Name     string `env:"NAME"  validate:"min=3"`
}
```

- Numbers: `Port` must be ≥ 1
- Strings: `Name` must be ≥ 3 characters

### `max=N`

Maximum value (for numbers) or maximum length (for strings).

```go
type Config struct {
    Port    int    `env:"PORT"  validate:"min=1,max=65535"`
    Comment string `env:"COMMENT" validate:"max=500"`
}
```

- Numbers: `Port` must be ≤ 65535
- Strings: `Comment` must be ≤ 500 characters

### `oneof=a b c`

String must match one of the space-separated options.

```go
type Config struct {
    LogLevel  string `env:"LOG_LEVEL" validate:"oneof=debug info warn error" default:"info"`
    Environment string `env:"ENV" validate:"oneof=dev staging prod"`
}
```

## Combining Rules

Rules are comma-separated and all must pass:

```go
type Config struct {
    Port     int    `env:"PORT" validate:"required,min=1,max=65535"`
    Username string `env:"USERNAME" validate:"required,min=3,max=64"`
    LogLevel string `env:"LOG_LEVEL" validate:"oneof=debug info warn error"`
}
```

Order doesn't matter — all rules are checked.

## Error Messages

When validation fails, `confkit.Explain(err)` shows clear, actionable errors:

```
Invalid configuration:

  Port
    error: must be between 1 and 65535, got 99999
    source: env (PORT)

  Username
    error: must be between 3 and 64 characters, got 2
    source: yaml (config.yaml)
```

## Validation Order

1. **Parse** — String values → typed values
2. **Validation** — Check rules from tags
3. **Defaults** — Apply `default` if field not set
4. **Interpolation** — Resolve `${VAR}` references

If validation fails, defaults are not applied and loading stops.

## Custom Validators

Define custom validation functions when built-in rules aren't enough:

```go
import (
    "fmt"
    "reflect"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Port int `env:"PORT" validate:"port-range"`
}

cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("port-range", func(v reflect.Value) error {
        n := v.Int()
        if n < 1024 || n > 49151 {
            return fmt.Errorf("must be a registered port (1024–49151), got %d", n)
        }
        return nil
    }),
)
```

### Using Multiple Custom Validators

```go
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("port-range", portValidator),
    confkit.WithValidator("valid-url", urlValidator),
    confkit.WithValidator("email", emailValidator),
)
```

Then use them in tags:

```go
type Config struct {
    Port  int    `env:"PORT"  validate:"port-range"`
    WebURL string `env:"URL"   validate:"valid-url"`
    Email string `env:"EMAIL" validate:"email"`
}
```

### Custom Validator Example: Regex

```go
import "regexp"

cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("hex-color", func(v reflect.Value) error {
        color := v.String()
        matched, _ := regexp.MatchString(`^#[0-9A-F]{6}$`, color)
        if !matched {
            return fmt.Errorf("must be a valid hex color (e.g. #FF0000), got %s", color)
        }
        return nil
    }),
)
```

## Nested Struct Validation

Validation rules apply to all nested structs:

```go
type Config struct {
    Server struct {
        Port int `env:"PORT" validate:"min=1,max=65535"`
        Host string `env:"HOST" validate:"required"`
    }
    Database struct {
        URL      string `env:"URL" validate:"required"`
        MaxConns int `env:"MAX_CONNS" validate:"min=1,max=1000"`
    } `prefix:"DB_"`
}
```

Each field is validated independently. Errors include the full path:

```
Invalid configuration:

  Server.Port
    error: must be between 1 and 65535, got 99999

  Database.URL
    error: field is required
```

## Conditional Validation (Custom)

For logic like "validate if another field is set", use custom validators:

```go
type Config struct {
    UseVault bool   `env:"USE_VAULT" default:"false"`
    VaultAddr string `env:"VAULT_ADDR"` // optional if UseVault=false
}

cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
    confkit.WithValidator("conditional-vault", func(v reflect.Value) error {
        // This validator receives the entire struct or field
        // Logic here to validate based on other field values
        return nil
    }),
)
```

## Skip Validation

For testing or specific scenarios, you can skip validation:

```go
// Load without validation (use with caution)
cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
)
// Validation still runs; there's no "skip" option
```

If you need to bypass validation, consider:
- Using a separate test struct without validation rules
- Loading into `map[string]interface{}` separately

## Type Coercion Before Validation

Values are parsed to their target types before validation:

```go
type Config struct {
    Port int `env:"PORT" validate:"min=1,max=65535"`
}

// env PORT="9999" → parsed to int 9999 → validated against min/max
```

If parsing fails, validation is never reached. Parse errors take precedence.

## Empty vs Zero Values

`required` rejects zero values:

```go
type Config struct {
    Port int `validate:"required"`  // Must be > 0
    Name string `validate:"required"` // Must be non-empty
    Enabled bool `validate:"required"` // Must be true
}
```

Use `default` to provide non-zero values:

```go
type Config struct {
    Port int `validate:"required,min=1,max=65535" default:"8080"`
    // If PORT not provided, uses 8080, passes validation
}
```

## Validation Errors Programmatically

Access validation errors as structured data:

```go
cfg, err := confkit.Load[Config](confkit.FromEnv())
if err != nil {
    report := err.(*confkit.ErrorReport)
    for _, fieldErr := range report.Errors {
        fmt.Printf("Field: %s\n", fieldErr.Path)
        fmt.Printf("Error: %s\n", fieldErr.Message)
        fmt.Printf("Rule: %s\n", fieldErr.Rule)
        fmt.Printf("Value: %s\n", fieldErr.Value) // redacted if Secret=true
    }
}
```

## Best Practices

1. **Use `required` for critical fields**
   ```go
   DatabaseURL string `env:"DATABASE_URL" validate:"required"`
   ```

2. **Combine min/max for ranges**
   ```go
   Port int `env:"PORT" validate:"min=1,max=65535"`
   ```

3. **Use `oneof` for enums**
   ```go
   Environment string `env:"ENV" validate:"oneof=dev staging prod"`
   ```

4. **Provide defaults for optional fields**
   ```go
   Timeout time.Duration `env:"TIMEOUT" default:"30s" validate:"min=1s,max=5m"`
   ```

5. **Use custom validators for complex logic**
   ```go
   Email string `env:"EMAIL" validate:"email"`  // custom validator
   ```

6. **Mark secrets with `secret:"true"`**
   ```go
   Password string `env:"DB_PASSWORD" validate:"required" secret:"true"`
   ```

## Next Steps

- **[Defaults](./defaults.md)** — Default values via struct tags
- **[Secret Redaction](./secret-redaction.md)** — Mark fields as sensitive
- **[Recipes](../recipes/)** — Real-world validation examples
