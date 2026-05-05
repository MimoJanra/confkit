# Changelog

All notable changes to confkit are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-05-05

### Added - Enterprise Secret Sources

- **HashiCorp Vault source** (`vault.FromVault*`)
  - Support for KV v1 and v2 engines
  - Multiple authentication methods: Token, AppRole, Kubernetes auth
  - Automatic secret loading and reload on changes

- **Consul KV source** (`consul.FromConsul*`)
  - Load configuration from Consul key-value store
  - Watch-based reload support
  - Token-based authentication

- **etcd v3 source** (`etcd.FromEtcd*`)
  - Load configuration from etcd with watch support
  - Multi-endpoint failover
  - Custom key prefix support

- **AWS Secrets Manager source** (`aws.FromAWSSecretsManager*`)
  - Load JSON/string secrets from AWS Secrets Manager
  - Multi-region failover support
  - Automatic secret refresh with TTL

### Added - Secret Management

- **Secret rotation**
  - Time-based and event-based rotation triggers
  - Rotation callback system for graceful config updates
  - In-flight request handling during rotation

- **Multi-region failover** (AWS sources)
  - Region health tracking and caching
  - Automatic failover to secondary regions
  - Region-specific configuration support

### Added - Production Features

- Hot reload with file watching (v0.4)
- Kubernetes ConfigMap source (v0.4)
- Config dumps with secret redaction (v0.4)
- Structured logging and observability hooks (v0.4)

### Documentation

- Complete iteration docs (Docs/01-iteration-v01-mvp.md through Docs/05-iteration-v05.md)
- Implementation roadmap with week-by-week breakdown
- API reference with all supported types
- Contributing guidelines

---

## [0.4.0] - 2026-04-15

### Added

- **Hot reload** with file watching
  - `LoadWithWatcher[T]()` API for automatic config reloading
  - File modification detection with configurable poll interval
  - Listener callback system for reload events

- **Kubernetes ConfigMap source**
  - Load from mounted ConfigMap volumes
  - In-cluster API support with service account authentication
  - Automatic reload on ConfigMap updates

- **AWS Systems Manager Parameter Store source**
  - Load configuration from AWS SSM Parameter Store
  - Parameter path hierarchy support (e.g., `/myapp/db/host`)
  - Region-specific parameter loading

- **Observability**
  - Config dump endpoint with secret redaction
  - Validation metrics and error tracking
  - Structured logging for load operations

---

## [0.3.0] - 2026-03-15

### Added

- **Environment variable prefix support**
  - `prefix:"APP_"` tag for struct-level env prefixing
  - Hierarchical prefixing for nested structs
  - Multi-level prefix concatenation

- **String interpolation**
  - `${VAR_NAME}` syntax in default values and loaded values
  - Environment variable and config field resolution
  - Circular reference detection

- **Functional options API**
  - `LoadWithOptions[T]()` for fine-grained control
  - `WithValidator()`, `WithSource()`, `WithMiddleware()` options
  - Custom source registry for extensibility

- **Command-line help text generation**
  - Auto-generated `--help` output from struct tags
  - Field descriptions, defaults, and validation rules in help
  - `short:"f"` tag support for single-letter flags
  - `hidden:"true"` tag to exclude fields from help

### Changed

- Refactored `Load()` to accept variadic sources (cleaner API)

---

## [0.2.0] - 2026-02-15

### Added

- **TOML source** (`confkit.FromTOML()`)
  - Load configuration from TOML files
  - Full TOML spec support (comments, arrays, tables)

- **Schema generation** (`schema.GenerateSchema[T]()`)
  - JSON Schema (draft-07 compatible) generation from structs
  - Markdown documentation generation
  - CLI help text generation

- **Custom validator registration**
  - Per-load custom validators without global state
  - Named validator functions with custom error messages

- **Improved documentation**
  - Schema generation examples
  - Custom validator examples
  - Comprehensive type support table

### Tests

- 38/38 integration tests passing
- Schema generation verified against standard validators

---

## [0.1.0] - 2026-01-15

### Added

- **Core configuration loading** (`Load[T]()`)
  - Type-safe config loading with Go generics
  - Multi-source support with precedence ordering
  - Default value support with late binding

- **Built-in sources**
  - Environment variables (`FromEnv()`)
  - YAML files (`FromYAML()`)
  - JSON files (`FromJSON()`)
  - Command-line flags (`FromFlags()`)

- **Validation system**
  - Built-in validators: `required`, `min`, `max`, `oneof`
  - Custom validation per field
  - Validation error aggregation

- **Human-readable error messages**
  - Structured error reporting with field context
  - Source attribution (which source caused the error)
  - Clear explanation of what failed and why

- **Secret redaction**
  - `secret:"true"` tag for sensitive fields
  - Automatic redaction in errors and logs
  - Safe error message formatting

- **Nested struct support**
  - Recursive struct traversal
  - Snake_case field name conversion
  - Type-safe nested configuration

- **Field metadata tracking**
  - FieldInfo abstraction for reflection
  - Struct tag parsing and caching
  - Source attribution for each field

### Tests

- 26/26 core tests passing
- 80%+ code coverage on core logic

---

## Roadmap

### v0.6 (Planned)

- Configuration composition / includes
- Multi-file YAML support (merge multiple files)
- Config templating (Helm-style)
- Performance optimizations

### v0.7 (Planned)

- Custom validation rules library (email, URL, IP, etc.)
- Conditional validation (Field B validation depends on Field A)
- Cross-field validation

### v0.8 (Planned)

- Prometheus metrics for config loads and reloads
- Structured tracing (OpenTelemetry)
- Config audit logging

### v1.0 (Planned)

- Public API stabilization
- Full semantic versioning
- Backward compatibility guarantees
- Official release announcement

---

## Comparison with Alternatives

| Feature | confkit | Viper | koanf |
|---------|---------|-------|-------|
| **Type Safety** | ✅ Generics `Load[T]` | ❌ String keys | ❌ Unmarshal |
| **Validation** | ✅ Built-in | ❌ Manual | ❌ Manual |
| **Error Messages** | ✅ Human-readable | ❌ Generic | ❌ Generic |
| **Secret Redaction** | ✅ Automatic | ❌ Manual | ❌ Manual |
| **Core Dependencies** | ✅ 2 packages | ❌ ~20 packages | ✅ Modular |
| **Enterprise Sources** | ✅ Optional modules | ❌ Bundled | ❌ Bundled |
| **String Interpolation** | ✅ `${VAR}` | ❌ | ❌ |
| **Hot Reload** | ✅ File watching | ✅ | ✅ |

---

## Contributors

- **Artem Alekseev** (@MimoJanra) — Creator and maintainer

---

## License

MIT License. See [LICENSE](LICENSE) for details.
