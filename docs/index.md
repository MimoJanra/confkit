---
layout: default
title: confkit — Typed Configuration for Go
---


# confkit

**Typed, validated configuration loading for Go** — the Pydantic equivalent for Go services.

Load configuration from multiple sources (environment, YAML, JSON, TOML, Kubernetes, AWS, Vault, Consul, etcd) with type safety, validation, and human-readable error messages.

## Quick Start

```go
type Config struct {
    Port int    `env:"PORT" default:"8080" validate:"min=1,max=65535"`
    DSN  string `env:"DATABASE_URL"        validate:"required" secret:"true"`
}

cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    confkit.FromYAML("config.yaml"),
)
if err != nil {
    log.Fatal(confkit.Explain(err))
}
```

**Output on validation error:**
```
Invalid configuration:

  DSN
    source: env
    error: field is required
```

## Why confkit?

| Feature | confkit | Viper | koanf |
|---|---|---|---|
| Typed return value | ✅ `Load[T]` | ❌ `GetString()` | ❌ `Unmarshal()` |
| Validation built-in | ✅ `validate` tag | ❌ manual | ❌ manual |
| Human-readable errors | ✅ field + source + value | ❌ | ❌ |
| Secret redaction | ✅ `secret:"true"` | ❌ | ❌ |
| String interpolation | ✅ `${VAR}` | ❌ | ❌ |

## Features by Version

- **v0.1-v0.5** ✅ Core features: loading, validation, defaults, cloud sources
- **v0.6-v0.9** ✅ Multi-file sources, 18+ built-in validators, observability (Prometheus, OpenTelemetry)
- **v1.0** ✅ API freeze, DevOps toolkit (Dump, ValidateOnly, LoadContext), production-ready

**Latest: v1.0.0** — Stable, production-ready with comprehensive examples, documentation, and API stability guarantee

## Install

```bash
go get github.com/MimoJanra/confkit@latest
```

## Resources

- **[Documentation]({{ '/docs/' | relative_url }})** — Full guides and tutorials
- **[API Reference]({{ '/api/' | relative_url }})** — Complete function and type reference
- **[Examples]({{ '/examples/' | relative_url }})** — Runnable code examples
- **[GitHub](https://github.com/MimoJanra/confkit)** — Source code

## License

MIT — see [LICENSE](https://github.com/MimoJanra/confkit/blob/main/LICENSE) file.
