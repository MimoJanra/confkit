package confkit

import (
	"confkit/tagutil"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func FromYAML(path string) Source {
	source, err := newYAMLSource(path)
	if err != nil {
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

func (y *yamlSource) Lookup(field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["yaml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := y.lookupNested(tagName, field.Path, field.AncestorTags)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// lookupNested navigates the yaml map using AncestorTags (yaml-first, then json/toml, then snake_case).
func (y *yamlSource) lookupNested(tagName, fieldPath string, ancestorTags []map[string]string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	current := any(y.data)

	for i := 0; i < len(parts)-1; i++ {
		key := ancestorKey(parts[i], i, ancestorTags, "yaml")
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

// ancestorKey returns the map key for ancestor level i, preferring preferredTag format.
func ancestorKey(fieldName string, i int, ancestorTags []map[string]string, preferredTag string) string {
	if i >= len(ancestorTags) || ancestorTags[i] == nil {
		return tagutil.SnakeCase(fieldName)
	}
	tags := ancestorTags[i]
	// Try preferred format first, then the other two, then snake_case
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
	return tagutil.SnakeCase(fieldName)
}
