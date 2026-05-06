package confkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

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
		cleanPath := filepath.Clean(path)
		raw, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", cleanPath, err)
		}
		m, err := parse(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", cleanPath, err)
		}
		deepMerge(merged, m)
	}
	return merged, nil
}

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

func (m *multiFileSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags[m.name]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		return "", false, nil
	}
	value, ok := lookupNested(m.data, tagName, field.Path, field.AncestorTags, m.name)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}
