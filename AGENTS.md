# Agent Instructions for confkit

confkit is a Go library for **typed, validated configuration loading** from multiple sources (YAML, JSON, TOML,
environment variables, Kubernetes, Vault, Consul, etcd, AWS SSM, AWS Secrets Manager).

## Project Intent

Use confkit when Go services need:

- **Type-safe config** via `Load[T]` generics (no `GetString()` accessors)
- **Validation at startup** (fail fast with clear error messages)
- **Automatic secret redaction** (fields tagged `secret:"true"` are never exposed in errors or logs)
- **Multiple sources** with explicit precedence (YAML → env vars → CLI flags)
- **Optional enterprise integrations** (Vault, Consul, etcd, AWS) without bloating the core package
- **Hot reload** support for config changes without restart

## Project Structure

```
confkit/
├── load.go              # Load[T], LoadWithOptions[T], LoadWithWatcher[T]
├── fieldinfo.go         # FieldInfo type and struct tag scanning
├── parser.go            # Type parsing (int, float, bool, duration, etc.)
├── sources.go           # Source interface
├── file_sources.go      # File format sources (YAML, JSON, TOML)
├── env_source.go        # FromEnv()
├── yaml_source.go       # FromYAML(path)
├── json_source.go       # FromJSON(path)
├── toml_source.go       # FromTOML(path)
├── flags_source.go      # FromFlags()
├── validation.go        # Built-in validators (required, min, max, oneof)
├── errors.go            # ErrorReport with field-level errors and redaction
├── options.go           # Functional options (WithSource, WithValidator, etc.)
├── interpolation.go     # String interpolation ${VAR}
├── watcher.go           # Hot reload / file watching
├── observability.go     # Config dump, metrics
├── k8s/                 # Kubernetes ConfigMap/Secret (separate module)
├── aws/                 # AWS SSM/Secrets (separate module)
├── vault/               # HashiCorp Vault (separate module)
├── consul/              # HashiCorp Consul KV (separate module)
├── etcd/                # etcd v3 (separate module)
├── aws/                 # AWS SSM + Secrets Manager (separate module)
└── examples/            # Runnable examples for each feature
```

## Important Commands

**Run tests:**

```bash
go test ./...
go test -cover ./...
```

**Format code:**

```bash
gofmt -w .
```

**Check module health:**

```bash
go mod tidy
go test ./...
```

**Run linter:**

```bash
golangci-lint run
```

## Code Style & Principles

- **Core is iron:** The five core pieces (loading, defaulting, validation, errors, types) must be rock solid. Don't add
  experimental features to core.
- **Keep core lightweight:** Only depend on `gopkg.in/yaml.v3` and `github.com/pelletier/go-toml/v2` in the main
  package.
- **Cloud SDKs belong in submodules:** Vault, Consul, etcd, AWS, Kubernetes integrations go in optional `confkit/vault`,
  `confkit/consul`, etc.
- **Preserve clear error messages:** This is the killer feature. Error messages must identify: field name, rule that
  failed, actual value (redacted if secret), and source.
- **Never expose secret-tagged fields:** Any field with `secret:"true"` must be redacted in errors, dumps, logs, and
  anywhere else.
- **No global state:** Validators and middleware are registered per Load[T] call, never globally.

## Public API Stability

Be very careful when modifying these (breaking changes require major version bump):

- `Load[T](...Source) (T, error)`
- `LoadWithOptions[T](...Option) (T, error)`
- `LoadWithWatcher[T](filePath string, ...Source) (T, *ConfigWatcher, error)`
- `Explain(err error) string`
- `Source` interface (`Name()`, `Lookup(*FieldInfo)`)
- `FieldInfo` type
- Validation tag semantics (`required`, `min`, `max`, `oneof`)
- Secret redaction behavior (fields tagged `secret:"true"` must always be redacted)

Internal types like `ErrorReport`, `FieldError` can be modified with better backwards-compat.

## Documentation Updates

When changing public behavior:

1. Update `README.md` with examples and explanations
2. Update `docs/` (on GitHub Pages) with guides and tutorials
3. Update `llms.txt` with new API functions or tag semantics
4. Add/update examples in `examples/` folder
5. Update relevant iteration doc in `Docs/` (e.g., `Docs/03-iteration-v03.md`)

## Testing Guidelines

- **Unit tests:** Each source, parser, and validator tested independently
- **Integration tests:** Multiple sources, precedence, nested structs, full workflows
- **Example tests:** Every documented example must run and produce expected output
- **Coverage target:** 80%+ on core logic

Write table-driven tests for validators and type parsing.

## Release Process

1. Bump version in `go.mod` (follows semantic versioning)
2. Update `CHANGELOG.md` with summary of changes
3. Create git tag: `git tag v0.X.Y`
4. Push tag: `git push origin v0.X.Y`
5. GitHub Actions builds and publishes release notes automatically

## LLM & AI Agent Patterns

confkit is well-suited for Go services that use LLMs:

**Pattern 1: LLM API client config**

```go
type LLMConfig struct {
    Provider   string  `env:"LLM_PROVIDER" validate:"oneof=claude,openai,anthropic"`
    APIKey     string  `env:"LLM_API_KEY" secret:"true" validate:"required"`
    Model      string  `env:"LLM_MODEL" default:"claude-3-5-sonnet"`
    Temperature float64 `env:"LLM_TEMPERATURE" default:"0.7" validate:"min=0,max=1"`
    MaxTokens  int     `env:"LLM_MAX_TOKENS" default:"4096"`
}

cfg, err := confkit.Load[LLMConfig](confkit.FromEnv())
```

**Pattern 2: Multi-model configuration**

```go
type Config struct {
    Primary   LLMConfig
    Fallback  LLMConfig
    CacheTTL  time.Duration `env:"CACHE_TTL" default:"1h"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("llm-config.yaml"),
    confkit.FromEnv(),
)
```

**Pattern 3: Hot reload for prompt templates or LLM parameters**

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config](
    "llm-config.yaml",
    confkit.FromYAML("llm-config.yaml"),
)

go func() {
    for newCfg := range watcher.Changes() {
        // Update LLM client with new parameters
        updateLLMClient(newCfg)
    }
}()
```

## Questions for Agent Implementation?

- How to add a new source type? → Implement `Source` interface, look at `env_source.go` as reference
- How to add custom validation? → Use `confkit.WithValidator("name", func(v reflect.Value) error { ... })`
- How to handle secrets safely? → Tag with `secret:"true"` and use `confkit.Explain()` for errors
- How to support nested config? → Use nested struct fields; optional `prefix` tag for env prefix
- How to debug config loading? → Call `confkit.Explain(err)` for human-readable errors

## Key Files to Reference

- `llms.txt` — LLM-readable API summary (this file is served at `https://mimojanra.github.io/confkit/llms.txt`)
- `README.md` — Project overview, quick start, examples
- `SECURITY.md` — Vulnerability reporting, security best practices
- `docs/` — Full documentation (Getting Started, API Reference, Examples)
- `Docs/01-iteration-v01-mvp.md` — v0.1 specification and design decisions
- `CONTRIBUTING.md` — Contribution guidelines
