## Description

Provide a brief description of the changes in this PR.

**Closes:** (link to related issue, e.g., #123)

## Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 📚 Documentation (changes to docs, README, examples)
- [ ] 🔧 Refactoring (code reorganization, no behavior change)
- [ ] ⚡ Performance improvement
- [ ] 🔐 Security fix

## What does this PR do?

Provide a more detailed description of the changes:

- What problem does it solve?
- What is the new behavior?
- Why is this approach the best?

### Before & After

If applicable, show how the API or behavior changes:

```go
// Before
cfg, err := confkit.Load[Config](/* ... */)

// After
cfg, err := confkit.LoadWithOptions[Config](
    confkit.WithSource(confkit.FromEnv()),
)
```

## Testing

How have you tested this change?

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Examples in docs tested
- [ ] Manual testing completed

### Test Coverage

Describe test cases that cover this change:

- Case 1: Happy path (what should work)
- Case 2: Error case (what should fail gracefully)
- Case 3: Edge case (any boundary conditions)

## Checklist

- [ ] Code follows confkit style (see CONTRIBUTING.md)
- [ ] Comments added for non-obvious logic
- [ ] Tests pass: `go test ./...`
- [ ] Coverage maintained: 80%+ on core logic
- [ ] README.md updated (if feature change)
- [ ] CHANGELOG.md updated (if feature/bug fix)
- [ ] Commit messages are clear and follow conventions
- [ ] No breaking changes to public API

## Breaking Changes?

If this PR includes breaking changes, describe them:

- [ ] No breaking changes
- [ ] Yes, breaking changes (describe below)

**Breaking change description:**

---

## Reviewers

Assign the maintainer for review: @MimoJanra
