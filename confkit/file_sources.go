package confkit

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlSource reads configuration from a YAML file.
type yamlSource struct {
	path string
	data map[string]any
}

// newYAMLSource creates a new YAML file source.
func newYAMLSource(path string) (*yamlSource, error) {
	s := &yamlSource{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load reads and parses the YAML file.
func (y *yamlSource) load() error {
	data, err := os.ReadFile(y.path)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	if err := yaml.Unmarshal(data, &y.data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if y.data == nil {
		y.data = make(map[string]any)
	}

	return nil
}

// Name returns the source name.
func (y *yamlSource) Name() string {
	return "yaml"
}

// Lookup retrieves a value from the YAML data.
func (y *yamlSource) Lookup(field *FieldInfo) (any, bool, error) {
	// Try yaml tag first, then json tag
	tagName := field.Tags["yaml"]
	if tagName == "" {
		tagName = field.Tags["json"]
	}
	if tagName == "" {
		return "", false, nil
	}

	// For top-level fields, use tag name directly
	// For nested fields, navigate using field path with tag names
	value, ok := y.lookupNested(tagName, field.Path)
	if !ok {
		return "", false, nil
	}

	// Return the value as-is (arrays stay as slices, scalars as strings or numbers)
	return value, true, nil
}

// lookupNested looks up a value using tag name, for now handles top-level only
func (y *yamlSource) lookupNested(tagName, fieldPath string) (any, bool) {
	// For v0.1: only handle top-level fields
	// TODO: handle nested structs like "server.port" -> server: { port: ... }
	parts := strings.Split(fieldPath, ".")
	if len(parts) > 1 {
		// Nested field - not yet supported in v0.1
		return nil, false
	}

	// Top-level: look up by tag name
	current := any(y.data)
	switch v := current.(type) {
	case map[string]any:
		val, ok := v[tagName]
		return val, ok
	default:
		return nil, false
	}
}


// jsonSource reads configuration from a JSON file.
type jsonSource struct {
	path string
	data map[string]any
}

// newJSONSource creates a new JSON file source.
func newJSONSource(path string) (*jsonSource, error) {
	s := &jsonSource{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load reads and parses the JSON file.
func (j *jsonSource) load() error {
	data, err := os.ReadFile(j.path)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	if err := json.Unmarshal(data, &j.data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if j.data == nil {
		j.data = make(map[string]any)
	}

	return nil
}

// Name returns the source name.
func (j *jsonSource) Name() string {
	return "json"
}

// Lookup retrieves a value from the JSON data.
func (j *jsonSource) Lookup(field *FieldInfo) (any, bool, error) {
	// Try json tag first, then yaml tag
	tagName := field.Tags["json"]
	if tagName == "" {
		tagName = field.Tags["yaml"]
	}
	if tagName == "" {
		return "", false, nil
	}

	// For top-level fields, use tag name directly
	value, ok := j.lookupNested(tagName, field.Path)
	if !ok {
		return "", false, nil
	}

	// Return the value as-is (arrays stay as slices, scalars as strings or numbers)
	return value, true, nil
}

// lookupNested looks up a value using tag name, for now handles top-level only
func (j *jsonSource) lookupNested(tagName, fieldPath string) (any, bool) {
	// For v0.1: only handle top-level fields
	// TODO: handle nested structs
	parts := strings.Split(fieldPath, ".")
	if len(parts) > 1 {
		// Nested field - not yet supported in v0.1
		return nil, false
	}

	// Top-level: look up by tag name
	current := any(j.data)
	switch v := current.(type) {
	case map[string]any:
		val, ok := v[tagName]
		return val, ok
	default:
		return nil, false
	}
}


