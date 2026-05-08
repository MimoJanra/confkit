package confkit

import "context"

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
