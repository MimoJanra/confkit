# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

---

## Philosophy

МИНИМАЛЬНОЕ ИСПОЛЬЗОВАНИЕ ВНЕШНИХ БИБЛИОТЕК, ПРИНЦИП РАЗУМНОСТИ. ЧИСТЫЙ КОД, ЧИСТАЯ СТРУКТУРА, КЛОН Pydantic

---

## Project Overview

**confkit** is a typed configuration loading library for Go. It solves the boilerplate of loading, merging, validating,
and explaining configuration errors.

**Status:** v1.0.2 released — fully production-ready with API stability guarantee.

**Key differentiator:** Human-readable error messages and secret redaction as first-class features.

**Target users:** Go services that want typed configuration without boilerplate (alternative to Viper, more
validation-focused).

---

## Essential Reading (In This Order)

1. **README.md** — Project pitch and quick example
2. **Docss/00-overview.md** — Vision, roadmap, and design principles
3. **Docss/IMPLEMENTATION_ROADMAP.md** — Current status and what is done
4. **Docss/NN-iteration-vX.md** — Spec for the iteration you are working on

All iterations v0.1–v1.0 are complete. Read the relevant iteration doc when extending or fixing a specific version.

---

## Development Approach

This project uses **iteration-based development** tied to semantic versions:

- **v0.1 (MVP):** ✅ Core loading, defaults, validation, error handling.
- **v0.2:** ✅ TOML, schema generation, custom validators.
- **v0.3:** ✅ Env prefix, string interpolation, functional options, custom sources, help text.
- **v0.4:** ✅ Hot reload, Kubernetes, AWS SSM, observability.
- **v0.5:** ✅ Vault, Consul, etcd, AWS Secrets Manager, secret rotation, multi-region.
- **v0.6–v0.9:** ✅ Multi-file sources, 18+ validators, Prometheus, OpenTelemetry, enhanced docs.
- **v1.0:** ✅ API freeze, DevOps toolkit (Dump, ValidateOnly, LoadContext), production release.
- **v1.1+:** Planned — config composition, advanced validation, additional cloud integrations.

Each iteration is fully specified in `Docss/NN-iteration-vX.md`. Read the relevant iteration doc before modifying
existing features.

**When adding a new iteration:**

- Reference the "Features in vX.X" section
- Build in the order listed under "Implementation Order"
- Use the "Example: First Proof of Concept" or "Example: Full vX.X Setup" as your target
- Verify against the "Success Criteria" checklist before claiming done

---

## Code Structure

```
confkit/
├── load.go                 # Core Load[T] API
├── fieldinfo.go            # FieldInfo type and reflection scanning
├── sources.go              # Source interface
├── file_sources.go         # YAML/JSON/TOML sources, multi-file, overlay, shared helpers
├── env_source.go           # EnvSource
├── flags_source.go         # FlagsSource
├── validation.go           # Validation engine + 19 built-in validators
├── errors.go               # ErrorReport and human-readable formatting
├── options.go              # Functional options (v0.3)
├── registry.go             # Custom source registry (v0.3)
├── watcher.go              # Hot reload / file watching (v0.4)
├── observability.go        # LoadMetrics, audit logging helpers (v0.4)
├── dump.go                 # Dump[T], DumpString[T], DumpYAML[T] — typed config serialization (v1.0)
├── discovery.go            # FindFile, FindSource, DefaultSearchDirs (v1.0)
├── rotation.go             # Secret rotation engine (v0.5)
├── doc.go                  # Package-level documentation
├── internal/
│   ├── parser/
│   │   └── parser.go       # Primitive type parsing (unexported)
│   └── interpolation/
│       └── interpolation.go # String interpolation — ${VAR} (unexported)
├── tests/                  # Black-box integration tests (package confkit_test)
├── schema/
│   └── schema.go           # JSON Schema + Markdown + CLI help generation (v0.2)
├── structtags/
│   └── tagutil.go          # Struct tag parsing utilities
├── otel/
│   ├── doc.go              # OTel package documentation
│   └── tracer.go           # OpenTelemetry Load wrappers (v0.7)
├── prometheus/
│   ├── doc.go              # Prometheus package documentation
│   └── metrics.go          # Prometheus metrics hook (v0.7)
├── k8s/
│   ├── go.mod              # K8s module
│   ├── doc.go              # K8s package documentation
│   └── k8s_source.go       # Kubernetes ConfigMap source (v0.4)
├── aws/
│   ├── go.mod              # AWS module
│   ├── doc.go              # AWS package documentation
│   ├── aws_ssm_source.go   # AWS SSM Parameter Store source (v0.4)
│   ├── aws_secrets_source.go # AWS Secrets Manager source (v0.5)
│   └── multiregion.go      # Multi-region failover (v0.5)
├── vault/
│   ├── go.mod              # Vault module
│   ├── doc.go              # Vault package documentation
│   └── vault_source.go     # HashiCorp Vault source (v0.5)
├── consul/
│   ├── go.mod              # Consul module
│   ├── doc.go              # Consul package documentation
│   └── consul_source.go    # Consul KV source (v0.5)
├── etcd/
│   ├── go.mod              # etcd module
│   ├── doc.go              # etcd package documentation
│   └── etcd_source.go      # etcd v3 source (v0.5)
├── examples/               # Usage examples
└── go.mod
```

---

## Commit Conventions

**These are strict:**

- **One-line messages only.** No multi-line commit bodies. Examples:
    - `Add Load function`
    - `Implement EnvSource`
    - `Fix validation error formatting`
    - `Add nested struct support`

- **No AI/Claude attribution.** Never add footers like `Co-Authored-By: Claude ...`. The human author owns the commit.

- **Local commits only.** Use `git commit` freely, but **never `git push`** unless explicitly asked. Commits stay in the
  local repository.

---

## Testing Strategy

- **Black-box tests** (`tests/`): All public API tests live in `package confkit_test`. Import the package as an external consumer; no access to unexported symbols.
- **Unit tests** (`internal/*/`): Parser and interpolation tested independently in their own packages.
- **Integration tests** (`tests/integration_test.go`): Multiple sources, precedence, nested structs, end-to-end loads.
- **Example-based tests** (`examples/`): Every example in README and Docs must run and produce expected output.

Run tests via:

```bash
go test ./...
```

Maintain 80%+ coverage on core logic. Cloud source tests use mocked APIs.

---

## Key Architectural Decisions

1. **FieldInfo abstraction:** Struct tags are parsed once into `FieldInfo`, then reused by all sources. This keeps
   sources simple.

2. **Source interface is minimal:** `Name() string` and `Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)`. This enables
   custom sources without knowing internal structure.

3. **Validation is custom:** No external validation library. Custom engine handles `required`, `min`, `max`, `oneof`.
   Full control over error messages.

4. **Errors are structured, not strings:** `ErrorReport` contains `[]FieldError` so errors can be programmatically
   inspected AND prettily formatted.

5. **Defaults are late-bound:** Applied after all sources, only if field wasn't set. Metadata tracks which fields came
   from where.

6. **Secrets marked in struct tags:** `secret:"true"` on the field means "redact everywhere" (errors, dumps, logs). This
   is checked globally, not per-source.

---

## Dependencies

Keep minimal. Core module depends only on:

- **gopkg.in/yaml.v3** — YAML parsing
- **github.com/pelletier/go-toml/v2** — TOML parsing (v0.2+)
- **Go stdlib only** for everything else in core

Cloud integrations (v0.4+) bring their own SDK dependencies. Enterprise sources (v0.5+) are optional imports that users
pull only if they need them.

---

## Common Workflows

### Starting a new iteration

1. Read `Docss/NN-iteration-vX.md` completely.
2. Check `Docss/IMPLEMENTATION_ROADMAP.md` for week-by-week breakdown.
3. Implement in the order listed under "Implementation Order" in the iteration doc.
4. Reference "Example: First Proof of Concept" as your done criterion.

### Adding a new type to parser

1. Add test case in `internal/parser/parser_test.go`.
2. Implement parser logic in `internal/parser/parser.go`.
3. Test round-trip: `string → parse → typed value → string` (if applicable).

### Adding a new source

1. Implement `Source` interface: `Name() string` and `Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)`.
2. Add tests for happy path and error cases.
3. Update examples in `README.md` and relevant `Docss/` file.

### Writing error messages

Remember: error messages are the killer feature. They should help users fix their config, not confuse them.

- ❌ Bad: `Key: 'Config.Port' Error: Field validation for 'Port' failed on the 'min' tag`
- ✅ Good: `PORT must be between 1 and 65535, got 99999`

Include: field name, rule that failed, actual value (redacted if secret), source (env/yaml/flag).

### Handling golangci-lint errors

**In test files:** Always explicitly ignore errors that you don't intend to handle:

- ❌ Bad: `defer os.Remove(tmpFile)` — golangci-lint errcheck will fail
- ✅ Good: `defer func() { _ = os.Remove(tmpFile) }()` — explicitly ignores the error

This applies to any cleanup operations in tests where errors don't affect the test outcome. Don't defer and ignore — wrap in a closure and assign to blank identifier.

---

## Design Principles

Refer back to these when making decisions:

1. **Core is iron:** The five core pieces (loading, defaulting, validation, errors, types) must be rock solid. Don't
   rush them.

2. **Not Viper:** We're typed, explicit, and validation-first. Not flexible and dynamic.

3. **Struct-first:** Configuration as struct definitions. Everything flows from the struct tags.

4. **Fail fast with help:** Validation happens eagerly. Errors are clear and actionable.

5. **Minimal core, extensible edges:** Core is 3-4 files. Customization and integrations live at the edges (v0.3+,
   v0.5+).

---

## When Stuck

1. **On API design:** Check the "Example: First Proof of Concept" or "Example: Full vX.X Setup" in the iteration doc.
   The API should match those examples.

2. **On error handling:** Read section 2.6 (v0.1) or 2.7 (v0.1) for examples. The rule: include context (source, field,
   rule), redact secrets, format for humans.

3. **On type support:** Check section 7 of the iteration doc for what types are in scope for that version.

4. **On architecture:** Refer to "Основные интерфейсы" (section 5) or whichever iteration's "Internal Architecture"
   section applies.

---

## Quick Reference

| What                          | Where                                      |
|-------------------------------|--------------------------------------------|
| Project vision                | `README.md` + `Docss/00-overview.md`       |
| Implementation status         | `Docss/IMPLEMENTATION_ROADMAP.md`          |
| v0.1 spec (core API)          | `Docss/01-iteration-v01-mvp.md`            |
| v0.2 spec (schema, TOML)      | `Docss/02-iteration-v02.md`                |
| v0.3 spec (DX, options)       | `Docss/03-iteration-v03.md`                |
| v0.4 spec (hot reload, cloud) | `Docss/04-iteration-v04.md`                |
| v0.5 spec (secrets, rotation) | `Docss/05-iteration-v05.md`                |
| v1.0 spec (API freeze, DevOps)| `Docss/10-iteration-v10.md`                |
| v1.1 spec (GCP/Azure/compose) | `Docss/11-iteration-v11.md`                |
| v1.2 spec (cross-field, SourceV2) | `Docss/12-iteration-v12.md`            |
| v2.0 spec (breaking: Source)  | `Docss/20-iteration-v20.md`                |
| API examples                  | Each iteration doc, "Example:" section     |
| Design decisions              | `CLAUDE.md` (this file) and iteration docs |
| Library philosophy            | `Docss/philosophy.md`                      |

---

## Repo-Specific Notes

- Go version: **1.25.11** (minimum version required by current dependencies)
- Module name: `github.com/MimoJanra/confkit`
- No vendor/ directory; use go mod
- No code generation; keep code hand-written
- Examples live in `examples/` or inline in Docs

---

## Remember

The core is complete. When modifying existing behaviour, check the relevant iteration spec first — the design decisions
are documented there. When adding new features (v0.6+), follow the same pattern: write the iteration doc first, then
implement.

Human-readable errors and secret redaction are the killer features. Protect them.
