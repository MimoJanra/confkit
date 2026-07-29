package confkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MimoJanra/confkit/structtags"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// fileSource is implemented by single-file sources to enable FromOverlay path detection.
type fileSource interface {
	Source
	sourcePath() string
}

// --- YAML ---

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
	data, err := os.ReadFile(path) // #nosec G304
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

func (y *yamlSource) Name() string       { return "yaml" }
func (y *yamlSource) sourcePath() string { return y.path }

func (y *yamlSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["yaml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "-" {
		return "", false, nil
	}
	if tagName != "" {
		if value, ok := lookupNested(y.data, tagName, field.Path, field.AncestorTags, "yaml"); ok {
			return value, true, nil
		}
	}
	snakeName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(y.data, snakeName, field.Path, field.AncestorTags, "yaml")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// --- JSON ---

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
	data, err := os.ReadFile(path) // #nosec G304
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

func (j *jsonSource) Name() string       { return "json" }
func (j *jsonSource) sourcePath() string { return j.path }

func (j *jsonSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["json"]
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "-" {
		return "", false, nil
	}
	if tagName != "" {
		if value, ok := lookupNested(j.data, tagName, field.Path, field.AncestorTags, "json"); ok {
			return value, true, nil
		}
	}
	snakeName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(j.data, snakeName, field.Path, field.AncestorTags, "json")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// --- TOML ---

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
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read TOML file: %w", err)
	}
	if err := toml.Unmarshal(data, &s.data); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return s, nil
}

func (t *tomlSource) Name() string       { return "toml" }
func (t *tomlSource) sourcePath() string { return t.path }

func (t *tomlSource) Lookup(_ context.Context, field *FieldInfo) (any, bool, error) {
	tagName := field.Tags["toml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "-" {
		return "", false, nil
	}
	if tagName != "" {
		if value, ok := lookupNested(t.data, tagName, field.Path, field.AncestorTags, "toml"); ok {
			return value, true, nil
		}
	}
	snakeName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(t.data, snakeName, field.Path, field.AncestorTags, "toml")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// --- Multi-file ---

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

// OverlayPath computes the overlay file path for a given base path and environment.
// "config.yaml" + "prod" → "config.prod.yaml"
func OverlayPath(basePath, env string) string {
	ext := filepath.Ext(basePath)
	return strings.TrimSuffix(basePath, ext) + "." + env + ext
}

// FromOverlay wraps base and merges an environment-specific overlay file on top.
// If the overlay file does not exist, base is returned unchanged without error.
// base must be created by FromYAML, FromJSON, or FromTOML (single-file sources).
func FromOverlay(base Source, env string) Source {
	fs, ok := base.(fileSource)
	if !ok {
		return base
	}
	overlayPath := OverlayPath(fs.sourcePath(), env)
	if _, err := os.Stat(overlayPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return base
		}
		return &errorSource{err: fmt.Errorf("overlay %s: %w", overlayPath, err)}
	}
	var over Source
	switch filepath.Ext(overlayPath) {
	case ".json":
		over = FromJSON(overlayPath)
	case ".toml":
		over = FromTOML(overlayPath)
	default:
		over = FromYAML(overlayPath)
	}
	return &overlaySource{base: base, over: over}
}

type overlaySource struct {
	base Source
	over Source
}

func (o *overlaySource) Name() string { return o.over.Name() + "+overlay" }

func (o *overlaySource) Lookup(ctx context.Context, field *FieldInfo) (any, bool, error) {
	if v, ok, err := o.over.Lookup(ctx, field); ok || err != nil {
		return v, ok, err
	}
	return o.base.Lookup(ctx, field)
}

func mergeFiles(paths []string, parse func([]byte) (map[string]any, error)) (map[string]any, error) {
	merged := make(map[string]any)
	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		raw, err := os.ReadFile(cleanPath) // #nosec G304
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
	if tagName == "-" {
		return "", false, nil
	}
	if tagName != "" {
		if value, ok := lookupNested(m.data, tagName, field.Path, field.AncestorTags, m.name); ok {
			return value, true, nil
		}
	}
	snakeName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(m.data, snakeName, field.Path, field.AncestorTags, m.name)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// --- Shared lookup helpers ---

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
