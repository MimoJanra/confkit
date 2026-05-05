---
name: Bug report
about: Report a bug to help us improve confkit
title: "[BUG] "
labels: bug
assignees: ''

---

## Describe the bug
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:

```go
// Minimal code example that reproduces the issue
type Config struct {
    // ...
}

cfg, err := confkit.Load[Config](/* ... */)
```

## Expected behavior
What should happen instead?

## Actual behavior
What actually happens? Include full error output using `confkit.Explain(err)`.

```
Invalid configuration:

  Field
    source: env
    error: field is required
```

## Environment

- Go version: (e.g., 1.24)
- confkit version: (e.g., v0.5.0)
- Sources used: (e.g., FromEnv, FromYAML, FromVault)
- OS: (Linux, macOS, Windows)

## Configuration

If relevant, share your config struct and sources (without secrets):

```go
type Config struct {
    Port int    `env:"PORT" default:"8080"`
    DB   string `env:"DATABASE_URL" validate:"required"`
}

cfg, err := confkit.Load[Config](
    confkit.FromYAML("config.yaml"),
    confkit.FromEnv(),
)
```

## Additional context
Add any other context about the problem here.
