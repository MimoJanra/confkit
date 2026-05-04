package confkit

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Load loads configuration from multiple sources into a typed struct.
// Sources are applied left-to-right; later sources override earlier ones.
// Defaults are applied after all sources.
// Validation is run after all values are set.
// Returns the typed struct or an ErrorReport if any errors occurred.
func Load[T any](sources ...Source) (T, error) {
	var cfg T

	// Scan fields from the config struct
	fields := ScanFields(cfg)

	// Prepare to collect errors
	report := &ErrorReport{}
	parser := NewParser()
	validator := NewValidator()

	// Track which fields were set by sources (vs. defaults)
	fieldValues := make(map[string]any)
	fieldSources := make(map[string]string)

	// Try each field across all sources
	// Later sources override earlier ones
	for _, field := range fields {
		var sourceErr error

		// Try sources left to right; later sources override earlier ones
		for _, source := range sources {
			value, ok, err := source.Lookup(&field)
			if err != nil {
				sourceErr = err
				continue
			}
			if ok {
				fieldValues[field.Path] = value
				fieldSources[field.Path] = source.Name()
			}
		}

		// If no source provided a value and there's an IO error, record it
		if _, found := fieldValues[field.Path]; !found && sourceErr != nil {
			report.AddError(FieldError{
				Path:    field.Path,
				Kind:    ErrorKindIO,
				Message: sourceErr.Error(),
			})
		}
	}

	// Apply defaults for fields not set by sources
	for _, field := range fields {
		if _, ok := fieldValues[field.Path]; !ok && field.HasDefault {
			fieldValues[field.Path] = field.Tags["default"]
			fieldSources[field.Path] = "default"
		}
	}

	// Parse and set all field values
	val := reflect.ValueOf(&cfg).Elem()
	if err := setStructFields(val, fields, fieldValues, fieldSources, parser, report); err != nil {
		return cfg, report
	}

	// Early return if there are parse errors
	if !report.IsEmpty() {
		return cfg, report
	}

	// Validate the config
	validationErrors := validator.ValidateConfig(cfg, fields)
	if !validationErrors.IsEmpty() {
		report.Errors = append(report.Errors, validationErrors.Errors...)
	}

	if !report.IsEmpty() {
		return cfg, report
	}

	return cfg, nil
}

// setStructFields sets all parsed values into the struct.
func setStructFields(val reflect.Value, fields []FieldInfo, values map[string]any, sources map[string]string, parser *Parser, report *ErrorReport) error {
	for _, field := range fields {
		rawVal, ok := values[field.Path]
		if !ok {
			continue // Not provided, not defaulted
		}

		// Convert to string for parsing if needed
		strVal := anyToString(rawVal)

		// Parse the value
		parsed, err := parser.Parse(strVal, field.Type)
		if err != nil {
			source := sources[field.Path]
			report.AddError(FieldError{
				Path:    field.Path,
				Source:  source,
				Kind:    ErrorKindParse,
				Message: err.Error(),
				Value:   strVal,
				Secret:  field.IsSecret,
			})
			continue
		}

		// Set the field value
		setFieldValue(val, field.Path, parsed)
	}

	return nil
}

// anyToString converts any value to string for parsing.
func anyToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// setFieldValue sets a value on a struct field by path (e.g., "Server.Port").
func setFieldValue(val reflect.Value, fieldPath string, value any) {
	parts := strings.Split(fieldPath, ".")
	current := val

	// Navigate to the parent of the target field
	for i := 0; i < len(parts)-1; i++ {
		field := current.FieldByName(parts[i])
		if !field.IsValid() {
			return
		}
		// Handle pointer dereference for nested structs
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		current = field
	}

	// Set the final field
	field := current.FieldByName(parts[len(parts)-1])
	if !field.IsValid() {
		return
	}
	field.Set(reflect.ValueOf(value))
}

// FromEnv creates an environment variable source.
func FromEnv() Source {
	return &envSource{}
}

// envSource is a local implementation of the environment variable source
type envSource struct{}

func (e *envSource) Name() string {
	return "env"
}

func (e *envSource) Lookup(field *FieldInfo) (any, bool, error) {
	envName := field.Tags["env"]
	if envName == "" {
		return "", false, nil
	}

	value, ok := os.LookupEnv(envName)
	return value, ok, nil
}

// FromYAML creates a YAML file source. Returns an error if the file cannot be loaded.
func FromYAML(path string) Source {
	source, err := newYAMLSource(path)
	if err != nil {
		// Return a source that always errors (will be caught during Load)
		return &errorSource{err: err}
	}
	return source
}

// FromJSON creates a JSON file source. Returns an error if the file cannot be loaded.
func FromJSON(path string) Source {
	source, err := newJSONSource(path)
	if err != nil {
		// Return a source that always errors (will be caught during Load)
		return &errorSource{err: err}
	}
	return source
}

// errorSource is a source that always returns an error (for deferred error handling).
type errorSource struct {
	err error
}

func (e *errorSource) Name() string {
	return "file"
}

func (e *errorSource) Lookup(_ *FieldInfo) (any, bool, error) {
	return "", false, e.err
}

// FromFlags creates a command-line flags source.
func FromFlags() Source {
	return &flagsSource{}
}

// flagsSource reads configuration from command-line flags (not yet implemented).
type flagsSource struct{}

func (f *flagsSource) Name() string {
	return "flag"
}

func (f *flagsSource) Lookup(_ *FieldInfo) (any, bool, error) {
	// TODO: implement flag parsing
	return "", false, nil
}
