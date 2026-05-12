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
// "/etc/app/config.toml" + "staging" → "/etc/app/config.staging.toml"
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
	if tagName != "" {
		value, ok := lookupNested(m.data, tagName, field.Path, field.AncestorTags, m.name)
		if ok {
			return value, true, nil
		}
	}

	snakeCaseTagName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(m.data, snakeCaseTagName, field.Path, field.AncestorTags, m.name)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}
