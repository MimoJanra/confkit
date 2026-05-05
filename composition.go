package confkit

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// FromYAMLFiles merges multiple YAML files. Later files override earlier ones.
func FromYAMLFiles(paths ...string) Source {
	merged, err := mergeFiles(paths, func(data []byte) (map[string]any, error) {
		var m map[string]any
		return m, yaml.Unmarshal(data, &m)
	})
	if err != nil {
		return &errorSource{err: err}
	}
	return &multiFileSource{name: "yaml", data: merged}
}

// FromJSONFiles merges multiple JSON files. Later files override earlier ones.
func FromJSONFiles(paths ...string) Source {
	merged, err := mergeFiles(paths, func(data []byte) (map[string]any, error) {
		var m map[string]any
		return m, json.Unmarshal(data, &m)
	})
	if err != nil {
		return &errorSource{err: err}
	}
	return &multiFileSource{name: "json", data: merged}
}

// FromTOMLFiles merges multiple TOML files. Later files override earlier ones.
func FromTOMLFiles(paths ...string) Source {
	merged, err := mergeFiles(paths, func(data []byte) (map[string]any, error) {
		var m map[string]any
		return m, toml.Unmarshal(data, &m)
	})
	if err != nil {
		return &errorSource{err: err}
	}
	return &multiFileSource{name: "toml", data: merged}
}

func mergeFiles(paths []string, parse func([]byte) (map[string]any, error)) (map[string]any, error) {
	merged := make(map[string]any)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		m, err := parse(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		deepMerge(merged, m)
	}
	return merged, nil
}

// deepMerge merges src into dst. Nested maps are merged recursively; scalars are overwritten.
func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				deepMerge(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

type multiFileSource struct {
	name string
	data map[string]any
}

func (m *multiFileSource) Name() string { return m.name }

func (m *multiFileSource) Lookup(field *FieldInfo) (any, bool, error) {
	tagName := field.Tags[m.name]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := multiFileLookup(m.data, tagName, field.Path, field.AncestorTags, m.name)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

func multiFileLookup(data map[string]any, tagName, fieldPath string, ancestorTags []map[string]string, preferredTag string) (any, bool) {
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
