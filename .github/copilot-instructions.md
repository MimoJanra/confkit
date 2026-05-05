# Copilot Instructions for confkit

This is confkit, a Go library for **typed, validated configuration loading** from multiple sources (YAML, JSON, TOML, environment variables, Kubernetes, Vault, Consul, etcd, AWS).

## Before Proposing Changes

- **Run tests first:** `go test ./...` must pass
- **Preserve the public API** — only break it if the issue explicitly requires a breaking change
- **Keep cloud/provider dependencies out of core** — they belong in optional submodules (`confkit/vault`, `confkit/consul`, etc.)
- **Never expose `secret:"true"` fields** — redact them in all error messages, dumps, and logs
- **Update docs when behavior changes** — README.md, docs/, examples/, llms.txt, and AGENTS.md must reflect the change

## Code Quality Rules

- **80%+ test coverage** on core logic (use `go test -cover ./...`)
- **Clear error messages** — include field name, source, rule that failed, and actual value (redacted if secret)
- **No global state** — validators and middleware are per-Load call, never globally registered
- **Minimal dependencies** — core only uses `gopkg.in/yaml.v3` and `github.com/pelletier/go-toml/v2`
- **Table-driven tests** for validators, parsers, and type conversions
- **Comment WHY, not WHAT** — code should be self-documenting; only comment hidden constraints or surprising behavior

## Testing Checklist

- [ ] Unit tests pass: `go test ./...`
- [ ] Coverage acceptable: `go test -cover ./...` shows 80%+ on touched files
- [ ] Example code runs: any new examples in README/docs are tested
- [ ] Integration tests work: multiple sources, precedence, nested structs
- [ ] No regressions: existing tests still pass
- [ ] Error messages are clear and redact secrets

## Documentation Checklist

- [ ] README.md updated (if public API changed)
- [ ] llms.txt updated (if API functions or tag semantics changed)
- [ ] docs/ updated (if behavior changed)
- [ ] examples/ updated (if patterns changed)
- [ ] AGENTS.md updated (if architectural decisions changed)

## Key Principles

1. **Core is iron** — keep loading, validation, errors, and secret redaction rock solid
2. **Type-safe first** — prefer generics and strong types over dynamic/string-based config
3. **Fail fast** — validate at startup, not after hours of debugging
4. **Human-readable errors** — error messages are the killer feature
5. **Secrets protected** — `secret:"true"` fields are never exposed anywhere
6. **Lean core** — enterprise features go in optional submodules

## File Structure

- `load.go` — Core Load[T], LoadWithOptions[T], LoadWithWatcher[T] functions
- `field_info.go` — FieldInfo type and struct tag scanning
- `parser.go` — Type parsing (int, float, bool, duration, slices, etc.)
- `sources.go` — Source interface definition
- `*_source.go` — Individual source implementations (env, yaml, json, toml, flags, k8s)
- `validation.go` — Validation engine and built-in validators
- `errors.go` — ErrorReport, FieldError, Explain() function
- `options.go` — Functional options (WithSource, WithValidator, WithMiddleware)
- `interpolation.go` — String interpolation ${VAR} support
- `watcher.go` — Hot reload / file watching
- `vault/`, `consul/`, `etcd/`, `aws/` — Optional enterprise sources (separate modules)
- `examples/` — Runnable examples for each feature
- `Docs/` — Iteration specs and design docs

## Do NOT

- Add feature flags or backwards-compat shims (just change the code)
- Add error handling for impossible scenarios (trust framework guarantees)
- Add TODO comments without a GitHub issue reference
- Validate input that should be validated by the caller (trust at boundaries)
- Export helper functions or types unless they're part of the public API contract
- Break the Source interface or Load[T] signature lightly

## Do

- Run `gofmt -w .` before committing
- Add test cases for edge cases
- Redact secrets in error messages
- Update docs when changing behavior
- Keep commits focused and clear
- Reference GitHub issues in commit messages when fixing bugs

## Questions?

See AGENTS.md for more detailed guidance on architecture, patterns, and LLM-specific use cases.
