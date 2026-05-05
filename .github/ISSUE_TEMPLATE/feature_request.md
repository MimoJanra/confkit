---
name: Feature request
about: Suggest an idea for confkit
title: "[FEATURE] "
labels: enhancement
assignees: ''

---

## Is your feature request related to a problem?
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

## Describe the solution you'd like
A clear and concise description of what you want to happen.

## Describe alternatives you've considered
A clear and concise description of any alternative solutions or features you've considered.

## Example use case

How would this feature be used in practice?

```go
type Config struct {
    // ...
}

cfg, err := confkit.Load[Config](
    // usage example with the new feature
)
```

## Alignment with confkit philosophy

Does this feature align with confkit's principles?

- [ ] Minimal dependencies (no bloat)
- [ ] Type-safe (leverages generics)
- [ ] Human-readable errors
- [ ] Secret redaction as first-class

## Additional context
Add any other context or screenshots about the feature request here.
