package confkit

import (
	"context"
	"os"
	"strings"
)

// FromEnv reads values from environment variables named by each field's `env`
// tag. Fields without that tag are skipped.
//
// The variable name is prefixed by the `prefix` tag of the field and of every
// enclosing struct, concatenated outermost first, so a `prefix:"APP_"` struct
// containing `env:"PORT"` reads APP_PORT.
func FromEnv() Source {
	return &envSource{}
}

type envSource struct{}

func (e *envSource) Name() string {
	return "env"
}

func (e *envSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	envName := field.Tags["env"]
	if envName == "" {
		return "", false, nil
	}

	prefix := buildEnvPrefix(field.AncestorTags)
	if p := field.Tags["prefix"]; p != "" {
		prefix += p
	}
	fullName := prefix + envName

	value, ok := os.LookupEnv(fullName)
	return value, ok, nil
}

func buildEnvPrefix(ancestorTags []map[string]string) string {
	var prefixes []string
	for _, tags := range ancestorTags {
		if p := tags["prefix"]; p != "" {
			prefixes = append(prefixes, p)
		}
	}
	return strings.Join(prefixes, "")
}

type errorSource struct {
	err error
}

func (e *errorSource) Name() string {
	return "error"
}

func (e *errorSource) Lookup(_ context.Context, _ *FieldInfo) (any, bool, error) {
	return "", false, e.err
}

// NewErrorSource returns a Source whose every lookup fails with err.
//
// It lets constructors such as FromYAML report a setup failure through the normal
// error report instead of panicking or returning a second value.
func NewErrorSource(err error) Source {
	return &errorSource{err: err}
}
