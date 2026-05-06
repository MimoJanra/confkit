package confkit

import "context"

type Source interface {
	Name() string
	Lookup(ctx context.Context, field *FieldInfo) (any, bool, error)
}
