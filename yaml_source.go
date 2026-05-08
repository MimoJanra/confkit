package confkit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MimoJanra/confkit/structtags"

	"gopkg.in/yaml.v3"
)

func FromYAML(path string) Source {
	source, err := newYAMLSource(path)
	if err != nil {
		return &errorSource{err: err}
	}
	return source
}

func FromYAMLOptional(path string) Source {
	source, err := newYAMLSource(path)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist) {
			return &emptySource{}
		}
		return &errorSource{err: err}
	}
	return source
}

type yamlSource struct {
	path string
	data map[string]any
}

func newYAMLSource(path string) (*yamlSource, error) {
	s := &yamlSource{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	if err := yaml.Unmarshal(data, &s.data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	return s, nil
}

func (y *yamlSource) Name() string {
	return "yaml"
}

func (y *yamlSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["yaml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName != "" {
		value, ok := lookupNested(y.data, tagName, field.Path, field.AncestorTags, "yaml")
		if ok {
			return value, true, nil
		}
	}

	snakeCaseTagName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(y.data, snakeCaseTagName, field.Path, field.AncestorTags, "yaml")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

func lookupNested(data map[string]any, tagName, fieldPath string, ancestorTags []map[string]string, preferredTag string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	current := any(data)

	for i := 0; i < len(parts)-1; i++ {
		key := ancestorKey(parts[i], i, ancestorTags, preferredTag)
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

func ancestorKey(fieldName string, i int, ancestorTags []map[string]string, preferredTag string) string {
	if i >= len(ancestorTags) || ancestorTags[i] == nil {
		return structtags.SnakeCase(fieldName)
	}
	tags := ancestorTags[i]
	order := []string{preferredTag}
	for _, t := range []string{"yaml", "json", "toml"} {
		if t != preferredTag {
			order = append(order, t)
		}
	}
	for _, t := range order {
		if v := tags[t]; v != "" {
			return v
		}
	}
	return structtags.SnakeCase(fieldName)
}
