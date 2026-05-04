package confkit

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func FromJSON(path string) Source {
	source, err := newJSONSource(path)
	if err != nil {
		return &errorSource{err: err}
	}
	return source
}

type jsonSource struct {
	path string
	data map[string]any
}

func newJSONSource(path string) (*jsonSource, error) {
	s := &jsonSource{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}
	if err := json.Unmarshal(data, &s.data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	return s, nil
}

func (j *jsonSource) Name() string {
	return "json"
}

func (j *jsonSource) Lookup(field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["json"]
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := j.lookupNested(tagName, field.Path, field.AncestorTags)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

func (j *jsonSource) lookupNested(tagName, fieldPath string, ancestorTags []map[string]string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	current := any(j.data)

	for i := 0; i < len(parts)-1; i++ {
		key := ancestorKey(parts[i], i, ancestorTags, "json")
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
