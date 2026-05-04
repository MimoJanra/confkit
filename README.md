# Сonfkit

**Typed configuration loading library for Go.**

A Go library that makes configuration loading, validation, and error reporting painless.

## Quick Example

```go
type Config struct {
    Port int    `env:"PORT" flag:"port" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL" validate:"required,url" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
    confkit.FromFlags(),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

## Features

- **Typed config**: Struct-based, no `GetString()` magic
- **Multiple sources**: YAML, JSON, environment, CLI flags
- **Validation**: Built-in rules with human-readable errors
- **Defaults**: Declared next to fields
- **Secrets**: Automatic redaction in errors
- **Nested structs**: Full support for complex configs
- **Extensible**: Custom validators and sources (v0.3+)

---

**Language**: Go 1.21+  
**Module**: `github.com/you/confkit` (to be determined)  
**License**: TBD
