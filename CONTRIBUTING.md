# Contributing to confkit

Thank you for your interest in contributing to confkit! This document provides guidelines and instructions for
contributors.

## Philosophy

confkit follows these principles:

- **Minimal external dependencies** — Use only essential libraries (YAML, TOML parsers). Cloud SDKs are optional
  submodules.
- **Clean code, clean structure** — Inspired by Pydantic, but Go-native.
- **Type-safe first** — Leverage Go generics for compile-time safety.
- **Human-readable errors** — Error messages should help users fix their config, not confuse them.
- **Secret redaction** — Sensitive data must be redacted everywhere (errors, logs, dumps).

## Code Structure

```
confkit/
├── load.go                 # Core Load[T] API
├── fieldinfo.go            # FieldInfo type and reflection scanning
├── parser.go               # Primitive type parsing
├── sources.go              # Source interface
├── file_sources.go         # File format sources (YAML, JSON, TOML)
├── env_source.go           # EnvSource
├── yaml_source.go          # YAMLSource
├── json_source.go          # JSONSource
├── toml_source.go          # TOMLSource
├── flags_source.go         # FlagsSource
├── validation.go           # Custom validation engine
├── errors.go               # ErrorReport and human-readable formatting
├── options.go              # Functional options
├── interpolation.go        # String interpolation
├── registry.go             # Custom source registry
├── watcher.go              # Hot reload / file watching
├── observability.go        # Config dump, metrics
├── k8s/                    # Kubernetes ConfigMap submodule (optional)
├── aws/                    # AWS SSM/Secrets Manager submodule (optional)
├── vault/                  # HashiCorp Vault submodule (optional)
├── consul/                 # Consul submodule (optional)
├── etcd/                   # etcd submodule (optional)
├── schema/                 # JSON Schema + documentation generation
├── structtags/             # Struct tag parsing utilities
└── examples/               # Usage examples
```

## Getting Started

1. **Fork and clone** the repository:
   ```bash
   git clone git@github.com:YOUR-USERNAME/confkit.git
   cd confkit
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature
   ```

3. **Make changes** following the code style below.

4. **Run tests**:
   ```bash
   go test ./...
   go test -cover ./...
   ```

5. **Commit with a clear message** (see Commit Conventions below).

## Commit Conventions

- **One-line messages only**. No multi-line bodies.
    - ✅ Good: `Add Load function`
    - ✅ Good: `Implement EnvSource`
    - ❌ Bad: `Add Load function and improve error handling`

- **Imperative mood** (present tense):
    - ✅ `Add` not `Added`
    - ✅ `Fix` not `Fixed`
    - ✅ `Implement` not `Implements`

- **No AI attribution**. Commits are authored by humans.
    - ❌ Never add `Co-Authored-By: Claude ...` footers

- **Atomic commits**. One logical change per commit.

## Code Style

### Naming

- **Functions**: `CamelCase` for exported, `camelCase` for private
- **Constants**: `UPPER_SNAKE_CASE` for exported, `camelCase` for private
- **Interfaces**: `Reader`, `Writer`, `Source` (descriptive, -er/-or endings)

### Comments

- **No unnecessary comments** — let code speak for itself.
- **Add comments only for the WHY** — hidden constraints, subtle invariants, workarounds.
- **Don't comment the WHAT** — well-named functions already do that.

Examples of good comments:

```go
// Circular references are detected during interpolation resolution.
// We use a Set to track visited variables and fail early.
func resolveInterpolation(value string) (string, error) { ... }

// Use reflection to scan struct tags once and reuse FieldInfo.
// This avoids re-scanning on every source lookup.
```

Examples of bad comments:

```go
// This function loads the config
func Load[T any](sources ...Source) (T, error) { ... }

// Increment the counter
counter++
```

### Testing

- **Unit tests** for each component (parser, source, validator).
- **Integration tests** combining multiple sources.
- **Example-based tests** — all README/Docs examples must run.
- **Target 80%+ coverage** on core logic.

Test file naming: `*_test.go` (Go convention).

Test structure:

```go
func TestEnvSourceLookup(t *testing.T) {
    tests := []struct {
        name    string
        env     map[string]string
        field   *FieldInfo
        wantOk  bool
        wantErr bool
    }{
        // cases...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Error Handling

- **Structured errors**: Return `*ErrorReport` with `[]FieldError` for config errors.
- **Include context**: Field name, source, rule, actual value (redacted if secret).
- **Help users fix**: Error messages should be actionable.

Bad:

```
Key: 'Config.Port' Error: Field validation for 'Port' failed on the 'min' tag
```

Good:

```
PORT must be between 1 and 65535, got 99999
  source: env
  rule: min=1
```

## Adding a New Source

1. Implement the `Source` interface:
   ```go
   type Source interface {
       Name() string
       Lookup(field *FieldInfo) (any, bool, error)
   }
   ```

2. Add tests:
    - Happy path (field found)
    - Field not found
    - Source errors (connection failures, etc.)

3. Update `README.md` with usage example.

4. Add to the appropriate iteration doc (`Docs/NN-iteration-vX.md`).

5. Commit with message: `Implement <SourceName>Source`

## Adding a New Type to Parser

1. Add a test case in `parser_test.go`:
   ```go
   func TestParseYourType(t *testing.T) {
       // test logic
   }
   ```

2. Implement parsing logic in `parser.go`.

3. Test round-trip: `string → parse → typed value → string`.

4. Update `README.md` "Supported Types" table.

5. Commit: `Add <TypeName> type support`

## Adding Validation Rules

1. Implement the rule in `validation.go`.
2. Add tests covering:
    - Valid values (should not error)
    - Invalid values (should error)
    - Edge cases
3. Update `README.md` "Validation" section.
4. Commit: `Add <rule-name> validation rule`

## Documentation

- All public functions should have godoc comments.
- Update `README.md` when adding features.
- Update relevant `Docs/NN-iteration-vX.md` when adding iteration features.
- Add examples to `examples/` for new features.

## Testing Before Submission

Run this checklist before submitting a PR:

```bash
# Run all tests
go test ./...

# Check coverage
go test -cover ./...

# Verify examples
go run examples/*.go

# Lint (if available)
golangci-lint run ./...

# Format code
go fmt ./...
```

## Pull Request Process

1. **Create a PR** with a clear title and description.
2. **Reference issues** if applicable: `Closes #123`
3. **Provide context** — what does this change do and why?
4. **Link examples** from README or Docs.
5. **Wait for review** — the maintainer will provide feedback.

### PR Title Format

- `feat: Add ...` (new feature)
- `fix: Fix ...` (bug fix)
- `docs: Update ...` (documentation)
- `refactor: Refactor ...` (code reorganization)
- `test: Add ...` (tests)

## Reporting Issues

If you find a bug or have a feature request:

1. **Check existing issues** to avoid duplicates.
2. **Provide context**:
    - What Go version are you using?
    - What sources/features are you using?
    - Minimal reproduction case
3. **Include error output** (use `confkit.Explain()` for config errors).
4. **Suggest a fix** if you have one.

## Questions?

- Open an issue with the `question` label.
- Check `CLAUDE.md` for design principles.
- Read the relevant `Docs/NN-iteration-vX.md` for context on that version.

---

**Thank you for contributing! 🎉**

confkit is built by the community, for the community.
