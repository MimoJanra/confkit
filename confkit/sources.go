package confkit

// Source is the interface for configuration sources (env, YAML, flags, etc).
type Source interface {
	// Name returns the source name for error messages.
	Name() string

	// Lookup attempts to find a value for the given field.
	// Returns (value, found, error). Value can be string or any parsed type.
	Lookup(field *FieldInfo) (any, bool, error)
}
