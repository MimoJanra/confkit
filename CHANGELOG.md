# Changelog

All notable changes to confkit are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2026-05-06

### Added

- **`map[string]V` field support** — parse map fields from `KEY1=val1,KEY2=val2` format; supports any scalar value type
- **`ErrorReport.Unwrap() []error`** — enables `errors.Is` / `errors.As` traversal on multi-field error reports
- **`AuditLogger` called on failure** — audit callback now fires on validation and source errors in addition to successful loads

### Changed

- **`Source.Lookup` signature** — `context.Context` added as first parameter (breaking change, intentional before v1.0 API freeze)
- **Source priority changed to first-wins** — first source to provide a value wins; later sources are only consulted for fields not yet set (previously last-wins)
- **`FlagsSource` properly implemented** — parses `--key=value`, `--key value`, `-k value`, `-k=value`, and boolean `--flag` forms
- **Embedded struct field promotion** — anonymous (embedded) struct fields are promoted without an extra path prefix
- **SnakeCase acronym handling fixed** — `DatabaseURL` → `database_url`, `HTTPServer` → `http_server`
- **Middleware error handling** — a middleware error now skips the field entirely instead of falling through to the next source
- **Schema generation deterministic** — map iteration is now sorted, producing stable JSON Schema output across runs
- **Code deduplication** — `lookupNested` / `ancestorKey` shared across YAML/JSON/TOML/file sources; single `splitPath` implementation used throughout
- **`setFieldValue` nil-pointer initialisation** — nil pointer fields are now initialised when setting a nested path

### Removed

- **`ErrorKindMissing` constant** — dead constant removed from the public API

### Tests

- Core coverage: 90.1%

---

## [0.8.2] - 2026-05-06

### Added

- **AI-optimized comparison pages** — SEO-friendly pages for common Go config library searches:
  - `/confkit-vs-viper` — Type-safe config with validation vs Viper's dynamic approach
  - `/confkit-vs-envconfig` — Full-featured config toolkit vs environment-variable-only library
  - `/confkit-vs-koanf` — Struct-first typing vs map-based modularity
  - `/best-go-config-library` — Decision guide comparing all four Go config libraries
  - `/go-config-validation` — Built-in validation rules (min, max, oneof, required)
  - `/go-secret-redaction` — Automatic secret masking in errors and logs
- **Accuracy fixes** — Corrected cloud integration comparison:
  - koanf: marked as `optional` (has separate provider modules) instead of `bundled`
  - envconfig: marked as `N/A` (no cloud sources) instead of `bundled`

### Changed

- **Code structure refactoring** — Improved Go naming conventions:
  - `field_info.go` → `fieldinfo.go` (single word per Go standard library style)
  - `sources_registry.go` → `registry.go` (shorter, clearer name)
  - `composition.go` → `file_sources.go` (explicit: file format merging)
  - `v04_integration_test.go` → `integration_test.go` (version not needed in filename)
  - `tagutil/` → `structtags/` (more descriptive package name)
- **Module organization** — Enterprise sources now properly separated:
  - `k8s/` is now a separate go module (like aws, vault, consul, etcd)
  - Added `k8s/go.mod`, `k8s/go.sum`, `k8s/doc.go`
  - Updated `go.work` to include all submodules
- **Documentation updates** — Fixed all references to renamed files across:
  - CLAUDE.md
  - CONTRIBUTING.md
  - AGENTS.md
  - Docss/code-review.md
  - Docss/roadmap-to-v1.0.md
  - .github/copilot-instructions.md

### Fixed

- **Import cycle** — Moved `Source` interface to core package (removed `sources/` package)
- **Kubernetes tests** — Moved K8s-specific integration tests to `k8s/k8s_source_test.go` to prevent circular dependencies

---

## [0.8.1] - 2026-05-05

### Fixed

- **AWS SSM nested field lookups** — Field path round-trip no longer lossy; cache stores by full parameter path to
  preserve Go field casing
- **AWS multi-region region parameter** — Region argument now passed to AWS config loader instead of being ignored
- **AWS multi-region error handling** — Preserves last region error when all sources fail (instead of swallowing it)
- **Rotation engine IsRotating() state** — Now correctly set to true during rotation and false on completion
- **Watcher SetPollInterval race condition** — Uses atomic.Value to safely track interval changes; no more silent
  failures when interval changed after Start()
- **Watcher Stop() double-call panic** — Added sync.Once guard for idempotent shutdown
- **Watcher listener concurrent modification race** — Deep copies listener slice to prevent append corruption
- **Interpolation regex compilation** — Moved regex to package level and converted environment to map for O(1) variable
  lookup
- **Field scanning append footgun** — Explicitly allocates ancestor tag slices to prevent shared backing array
  corruption in nested structs
- **Kubernetes ConfigMap source** — Tag-based key resolution (env/yaml/json tags) with exact field name and snake_case
  fallbacks
- **Variable naming clarity** — Renamed inverted `interpolationErrors` flag to `interpolationOK`
- **Middleware error handling** — Continues to next source on middleware failure instead of breaking
- **Multi-region SSM path normalization** — Path prefix now normalized consistently with single-region helper

---

## [0.8.0] - 2026-05-05

### Added — Observability

- **Audit logging** (`WithAuditLogger`)
    - Callback receives `[]AuditEntry` after every successful load
    - Each entry has `Field`, `Source`, and `Value` (secrets redacted)
- **Load hooks** (`WithLoadHook`)
    - Low-level hook called after every load with `success bool`, `duration`, `errCount`
    - Used by Prometheus and OTel submodules
- **Prometheus submodule** (`confkit/prometheus`)
    - `NewMetrics(reg)` registers `confkit_loads_total`, `confkit_load_duration_seconds`, `confkit_errors_total`
    - `Metrics.Hook()` returns a `confkit.Option` — drop-in, no wrapping needed
- **OpenTelemetry submodule** (`confkit/otel`)
    - `otel.Load[T](ctx, tracer, sources...)` — wraps Load with a span
    - `otel.LoadWithOptions[T](ctx, tracer, options...)` — same for options form
    - Span attributes: `confkit.sources`, `confkit.success`

---

## [0.7.0] - 2026-05-05

### Added — Advanced Validation

- **18 built-in format validators** (no external dependencies):
    - `email` — valid email address
    - `url` — valid URL (any scheme)
    - `http_url` — valid HTTP or HTTPS URL
    - `ip` — valid IPv4 or IPv6 address
    - `ipv4` — valid IPv4 address
    - `ipv6` — valid IPv6 address
    - `uuid` — valid UUID (v1–v5)
    - `hostname` — valid hostname (RFC 1123)
    - `port` — valid port number (1–65535), works on int and string fields
    - `regex=pattern` — value must match the given regular expression
    - `len=N` — string must be exactly N characters (Unicode-aware)
    - `contains=str` — string must contain substring
    - `startswith=str` — string must start with prefix
    - `endswith=str` — string must end with suffix
    - `alpha` — letters only
    - `alphanum` — letters and digits only
    - `numeric` — digits only
    - `lowercase` — all lowercase
    - `uppercase` — all uppercase
    - `notempty` — must not be blank (non-whitespace)

- **Model validators** (`WithModelValidator[T](fn func(*T) error)`)
    - Cross-field validation: runs after all field validators pass
    - Receives a pointer to the fully populated config struct
    - Multiple model validators can be registered; all run independently

---

## [0.6.0] - 2026-05-05

### Added — Config Composition

- **Multi-file YAML** (`FromYAMLFiles(paths ...string)`)
    - Merges multiple YAML files; later files override earlier ones
    - Nested maps are merged recursively (deep merge)
- **Multi-file JSON** (`FromJSONFiles(paths ...string)`)
    - Same semantics as `FromYAMLFiles` for JSON
- **Multi-file TOML** (`FromTOMLFiles(paths ...string)`)
    - Same semantics as `FromYAMLFiles` for TOML

**Usage pattern:**

```go
cfg, err := confkit.Load[Config](
    confkit.FromYAMLFiles(
        "config/base.yaml",      // defaults
        "config/production.yaml", // env-specific overrides
        "config/local.yaml",      // developer local overrides
    ),
    confkit.FromEnv(),
)
```

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

### v1.0 (Planned)

- Public API stabilization
- Full semantic versioning
- Backward compatibility guarantees
- Official release announcement

---

## Comparison with Alternatives

| Feature                  | confkit              | Viper          | koanf       |
|--------------------------|----------------------|----------------|-------------|
| **Type Safety**          | ✅ Generics `Load[T]` | ❌ String keys  | ❌ Unmarshal |
| **Validation**           | ✅ Built-in           | ❌ Manual       | ❌ Manual    |
| **Error Messages**       | ✅ Human-readable     | ❌ Generic      | ❌ Generic   |
| **Secret Redaction**     | ✅ Automatic          | ❌ Manual       | ❌ Manual    |
| **Core Dependencies**    | ✅ 2 packages         | ❌ ~20 packages | ✅ Modular   |
| **Enterprise Sources**   | ✅ Optional modules   | ❌ Bundled      | ❌ Bundled   |
| **String Interpolation** | ✅ `${VAR}`           | ❌              | ❌           |
| **Hot Reload**           | ✅ File watching      | ✅              | ✅           |

---

## Contributors

- **Artem Alekseev** (@MimoJanra) — Creator and maintainer

---

## License

MIT License. See [LICENSE](LICENSE) for details.
