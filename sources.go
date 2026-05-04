package confkit

type Source interface {
	Name() string
	Lookup(field *FieldInfo) (any, bool, error)
}
