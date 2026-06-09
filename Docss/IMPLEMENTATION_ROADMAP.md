# Implementation Roadmap

Current status of confkit features by version.

---

## v0.1 — MVP (Complete ✅)

**Core loading, defaults, validation, error handling.**

- ✅ `Load[T]` basic API
- ✅ Environment variables (`FromEnv`)
- ✅ YAML files (`FromYAML`)
- ✅ JSON files (`FromJSON`)
- ✅ Struct tag parsing (env, default, validate)
- ✅ Type parsing (int, float, bool, string, duration)
- ✅ Validation (required, min, max, oneof)
- ✅ Error formatting (human-readable ErrorReport)
- ✅ Nested struct support
- ✅ Test coverage 85%+

---

## v0.2 — Schema & TOML (Complete ✅)

**TOML parsing, JSON Schema generation, custom validators.**

- ✅ TOML files (`FromTOML`)
- ✅ JSON Schema generation (`schema.GenerateSchema`)
- ✅ Markdown schema export
- ✅ CLI help text generation
- ✅ Custom validator registration
- ✅ Validator examples and docs

---

## v0.3 — DX & Options (Complete ✅)

**Environment prefix, string interpolation, functional options.**

- ✅ `prefix` struct tag for nested configs
- ✅ String interpolation (`${VAR}` syntax)
- ✅ `LoadWithOptions` functional API
- ✅ Custom source registry
- ✅ Context support (`LoadContext`, `WithContext`)
- ✅ Secret redaction infrastructure
- ✅ Help text in struct tags

---

## v0.4 — Cloud & Hot Reload (Complete ✅)

**Hot reload, Kubernetes, AWS SSM, observability.**

- ✅ File watching (`LoadWithWatcher`, `ConfigWatcher`)
- ✅ Kubernetes ConfigMap source (`k8s/k8s_source.go`)
- ✅ AWS SSM Parameter Store (`aws/aws_ssm_source.go`)
- ✅ Config dump (`DumpConfig`)
- ✅ Metrics infrastructure
- ✅ Audit logging

---

## v0.5 — Enterprise Secrets (Complete ✅)

**Vault, Consul, etcd, AWS Secrets Manager, secret rotation.**

- ✅ HashiCorp Vault source
  - ✅ Token auth
  - ✅ AppRole auth
  - ✅ Kubernetes auth
- ✅ Consul KV source
- ✅ etcd v3 source
- ✅ AWS Secrets Manager (`aws/aws_secrets_source.go`)
- ✅ Multi-region failover
- ✅ Secret rotation engine
- ✅ Rotation metrics

---

## v0.6 — Multi-File & Validators (Complete ✅)

**Multiple file sources, 19+ built-in validators.**

- ✅ Multi-file support (`FromYAMLFiles`, `FromJSONFiles`, `FromTOMLFiles`)
- ✅ 19+ validators:
  - ✅ Format validators (email, url, http_url, ip, ipv4, ipv6, uuid, hostname)
  - ✅ Length validator (len — exact character count)
  - ✅ Pattern validators (regex, alpha, alphanum, numeric, lowercase, uppercase)
  - ✅ String content validators (contains, startswith, endswith, notempty)
  - ✅ Number validators (min, max, port)
  - ✅ Enum validators (oneof)
- ✅ Validator composition
- ✅ Custom validator examples

---

## v0.7 — Observability (Complete ✅)

**Prometheus metrics, tracing integration.**

- ✅ Prometheus metrics (`prometheus/` module)
  - ✅ Load duration histogram (`confkit_load_duration_seconds`)
  - ✅ Validation error counter (`confkit_errors_total`)
  - ✅ Load success/failure counter (`confkit_loads_total`)
- ✅ OpenTelemetry instrumentation (`otel/` module)
- ✅ Structured logging support
- ✅ Audit log formatting

---

## v0.8 — Enhanced Docs (Complete ✅)

**Comprehensive documentation, examples, guides.**

- ✅ Full API documentation
- ✅ Getting started guide
- ✅ Configuration sources guide
- ✅ Validation rules guide
- ✅ Cloud integration recipes (Kubernetes, AWS, Vault, Consul, etcd)
- ✅ CLI flags guide
- ✅ Hot reload guide
- ✅ Secret rotation guide
- ✅ Runnable examples with tests

---

## v0.9 — Bug Fixes & Polish (Complete ✅)

**Bug fixes, performance tuning, documentation polish.**

- ✅ Error path handling improvements
- ✅ Context deadline propagation
- ✅ Memory efficiency optimizations
- ✅ Documentation updates
- ✅ Example enhancements
- ✅ Test coverage improvements (90%+)

---

## v1.0 — API Freeze & DevOps Toolkit (Complete ✅)

**API stability guarantee, DevOps features.**

- ✅ API stability contract (STABLE.md)
- ✅ `LoadContext[T]` — deadline support for cloud sources
- ✅ `ValidateOnly[T]` — dry-run in CI without side effects
- ✅ `Dump[T]` — typed config serialization (JSON/YAML)
- ✅ `DumpString[T]` — config dumps for logging
- ✅ `FindFile` — search config file in standard paths
- ✅ `FromOverlay` — base + environment config merging
- ✅ `ConfigDelta` — diff between old and new config
- ✅ `AddDeltaListener` — react to specific config changes
- ✅ `WithModelValidator[T]` — typed cross-field validation via Go function
- ✅ Production examples
- ✅ Migration guide from v0.9
- ✅ API stability documentation

---

## v1.1 — Cloud Completeness + Composition Primitives (Planned)

**Additive only. No breaking changes.**

- ⬜ GCP Secret Manager submodule (`gcp/`)
- ⬜ Azure Key Vault submodule (`azure/`)
- ⬜ `FromStruct[T]` / `FromStructAll[T]` — pre-loaded struct as a config source
- ⬜ Enhanced interpolation: bash-style `${VAR:-default}` and `${VAR:+alt}`

---

## v1.2 — Advanced Validation + Source Transition (Planned)

**Additive only. Deprecation notices for v2.0.**

- ⬜ Tag-based cross-field validators (`requiredIf`, `requiredWith`, `requiredWithAll`, `excludedIf`, `excludedWith`)
- ⬜ `FromDotEnv(path)` / `FromDotEnvAuto()` — `.env` file format
- ⬜ `FieldError.Dependents` — cross-field error context
- ⬜ `SourceV2` interface introduced (Lookup returns `string`) alongside deprecated `Source`
- ⬜ `WrapLegacySource` migration adapter
- ⬜ All built-in sources implement `SourceV2`

---

## v2.0 — Breaking: New Source Contract (Planned)

**First major release. Breaking changes to `Source` interface.**

- ⬜ `Source.Lookup` returns `(string, bool, error)` — was `(any, bool, error)`
- ⬜ `SourceV2` renamed to `Source`; old `Source` removed
- ⬜ `WrapLegacySource` removed
- ⬜ `FieldInfo.SourceName` — tracks which source provided each field value
- ⬜ Import path updated to `/v2`
- ⬜ `STABLE.md` updated, `MIGRATION_V2.md` written

---

## Feature Completion Summary

| Category | Status | Coverage |
|----------|--------|----------|
| Core API | ✅ Complete | 100% |
| File sources | ✅ Complete | 100% |
| Environment | ✅ Complete | 100% |
| Cloud sources (AWS, Vault, Consul, etcd, K8s) | ✅ Complete | 100% |
| Cloud sources (GCP, Azure) | ⬜ Planned v1.1 | — |
| Validation (per-field) | ✅ Complete | 100% |
| Validation (cross-field) | ⬜ Planned v1.2 | — |
| Error handling | ✅ Complete | 100% |
| Observability | ✅ Complete | 100% |
| Documentation | ✅ Complete | 100% |
| Test coverage | ✅ 90%+ | Production-grade |

---

## Current Version: v1.0.2

**Production ready.** All core features complete. API stable.

**API Stability Guarantee.** No breaking changes within v1.x series.

**Fully Documented.** API reference, guides, recipes, examples.

**Thoroughly Tested.** 90%+ test coverage. Real-world examples.

### v1.0.2 patch changes

- Security: Go raised to 1.25.11 (fixes GO-2026-5037, GO-2026-5039)
- Internal packages: `parser` and `interpolation` moved to `internal/`
- File consolidation: YAML/JSON/TOML sources merged into `file_sources.go`; built-in validators merged into `validation.go`
- Tests reorganized into `tests/` as black-box `package confkit_test`

---

## To Get Started

See the [Project Overview](./00-overview.md) for vision and architecture, or jump to:

- **[Installation](./installation.md)**
- **[Getting Started](./getting-started.md)**
- **[API Reference](../api/)**
