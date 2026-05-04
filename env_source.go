package confkit

import "os"

func FromEnv() Source {
	return &envSource{}
}

type envSource struct{}

func (e *envSource) Name() string {
	return "env"
}

func (e *envSource) Lookup(field *FieldInfo) (any, bool, error) {
	envName := field.Tags["env"]
	if envName == "" {
		return "", false, nil
	}
	value, ok := os.LookupEnv(envName)
	return value, ok, nil
}

type errorSource struct {
	err error
}

func (e *errorSource) Name() string {
	return "file"
}

func (e *errorSource) Lookup(_ *FieldInfo) (any, bool, error) {
	return "", false, e.err
}
