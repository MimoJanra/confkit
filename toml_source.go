package confkit

import (
	"context"
	"fmt"
	"os"

	"github.com/MimoJanra/confkit/structtags"
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
	if tagName != "" {
		value, ok := lookupNested(t.data, tagName, field.Path, field.AncestorTags, "toml")
		if ok {
			return value, true, nil
		}
	}

	snakeCaseTagName := structtags.SnakeCase(field.Name)
	value, ok := lookupNested(t.data, snakeCaseTagName, field.Path, field.AncestorTags, "toml")
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}
