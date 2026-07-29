package confkit

import "context"

// Source supplies raw configuration values for individual fields.
//
// Name identifies the source in error messages and audit entries ("env", "yaml").
// Lookup reports the value for field, whether it was present, and any retrieval
// error. A missing value must return (nil, false, nil) rather than an error, so
// that later sources and defaults can still apply.
//
// Implementations are consulted in the order given to Load, and the first source
// reporting found wins.
type Source interface {
	Name() string
	Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)
}

type emptySource struct{}

func (e *emptySource) Name() string {
	return "empty"
}

func (e *emptySource) Lookup(_ context.Context, _ *FieldInfo) (any, bool, error) {
	return nil, false, nil
}
