package confkit

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

func FromTOML(path string) Source {
	source, err := newTOMLSource(path)
	if err != nil {
		return &errorSource{err: err}
	}
	return source
}

type tomlSource struct {
	path string
	data map[string]any
}

func newTOMLSource(path string) (*tomlSource, error) {
	s := &tomlSource{path: path, data: make(map[string]any)}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read TOML file: %w", err)
	}
	if err := toml.Unmarshal(data, &s.data); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return s, nil
}

func (t *tomlSource) Name() string {
	return "toml"
}

func (t *tomlSource) Lookup(field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["toml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := t.lookupNested(tagName, field.Path, field.AncestorTags)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

func (t *tomlSource) lookupNested(tagName, fieldPath string, ancestorTags []map[string]string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	current := any(t.data)

	for i := 0; i < len(parts)-1; i++ {
		key := ancestorKey(parts[i], i, ancestorTags, "toml")
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}

	if m, ok := current.(map[string]any); ok {
		if v, found := m[tagName]; found {
			return v, true
		}
	}
	return nil, false
}
