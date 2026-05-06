---
layout: default
title: confkit — Typed Configuration for Go
---

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "confkit",
  "description": "Type-safe configuration toolkit for Go. Load, merge, validate, and redact configuration from structs, files, environment variables, CLI flags, and secret stores like Vault, AWS Secrets Manager, Consul, and etcd.",
  "url": "https://mimojanra.github.io/confkit/",
  "applicationCategory": "DeveloperApplication",
  "programmingLanguage": "Go",
  "codeRepository": "https://github.com/MimoJanra/confkit",
  "license": "https://opensource.org/license/mit",
  "runtimePlatform": "Go 1.21+",
  "softwareVersion": "0.5.0",
  "author": {
    "@type": "Organization",
    "name": "confkit Contributors"
  },
  "operatingSystem": ["Linux", "macOS", "Windows"],
  "keywords": "configuration, Go, config loader, type-safe, validation, secrets, Vault, AWS",
  "offers": {
    "@type": "Offer",
    "price": "0",
    "priceCurrency": "USD"
  },
  "image": "https://img.shields.io/badge/go-1.22%2B-blue",
  "downloadUrl": "https://github.com/MimoJanra/confkit/releases",
  "fileFormat": "application/x-golang"
}
</script>

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
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
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

- **v0.1** ✅ Core loading, defaults, validation, error handling
- **v0.2** ✅ TOML, schema generation, custom validators
- **v0.3** ✅ Env prefix, string interpolation, functional options, custom sources
- **v0.4** ✅ Hot reload, Kubernetes, AWS SSM, observability
- **v0.5** ✅ Vault, Consul, etcd, AWS Secrets Manager, secret rotation

## Install

```bash
go get github.com/MimoJanra/confkit@v0.5.0
```

## Resources

- **[Documentation]({{ '/docs/' | relative_url }})** — Full guides and tutorials
- **[API Reference]({{ '/api/' | relative_url }})** — Complete function and type reference
- **[Examples]({{ '/examples/' | relative_url }})** — Runnable code examples
- **[GitHub](https://github.com/MimoJanra/confkit)** — Source code

## License

MIT — see [LICENSE](https://github.com/MimoJanra/confkit/blob/main/LICENSE) file.
