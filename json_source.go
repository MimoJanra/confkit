package confkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func (j *jsonSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["json"]
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := lookupNested(j.data, tagName, field.Path, field.AncestorTags, "json")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}
