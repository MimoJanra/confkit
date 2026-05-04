package confkit

import (
	"fmt"
	"reflect"
	"strings"
)

// Load loads configuration into T from multiple sources.
// Sources are evaluated left-to-right; later sources override earlier ones.
func Load[T any](sources ...Source) (T, error) {
	var cfg T

	fields := ScanFields(cfg)
	report := &ErrorReport{}
	parser := NewParser()
	validator := NewValidator()

	fieldValues := make(map[string]any)
	fieldSources := make(map[string]string)

	for _, field := range fields {
		var sourceErr error
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
		if _, found := fieldValues[field.Path]; !found && sourceErr != nil {
			report.AddError(FieldError{
				Path:    field.Path,
				Kind:    ErrorKindIO,
				Message: sourceErr.Error(),
			})
		}
	}

	for _, field := range fields {
		if _, ok := fieldValues[field.Path]; !ok && field.HasDefault {
			fieldValues[field.Path] = field.Tags["default"]
			fieldSources[field.Path] = "default"
		}
	}

	val := reflect.ValueOf(&cfg).Elem()
	setStructFields(val, fields, fieldValues, fieldSources, parser, report)

	if !report.IsEmpty() {
		return cfg, report
	}

	validationErrors := validator.ValidateConfig(cfg, fields)
	if !validationErrors.IsEmpty() {
		report.Errors = append(report.Errors, validationErrors.Errors...)
	}

	if !report.IsEmpty() {
		return cfg, report
	}

	return cfg, nil
}

func setStructFields(val reflect.Value, fields []FieldInfo, values map[string]any, sources map[string]string, parser *Parser, report *ErrorReport) {
	for _, field := range fields {
		rawVal, ok := values[field.Path]
		if !ok {
			continue
		}

		strVal := anyToString(rawVal)
		parsed, err := parser.Parse(strVal, field.Type)
		if err != nil {
			report.AddError(FieldError{
				Path:    field.Path,
				Source:  sources[field.Path],
				Kind:    ErrorKindParse,
				Message: err.Error(),
				Value:   strVal,
				Secret:  field.IsSecret,
			})
			continue
		}

		setFieldValue(val, field.Path, parsed)
	}
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func setFieldValue(val reflect.Value, fieldPath string, value any) {
	parts := strings.Split(fieldPath, ".")
	current := val

	for i := 0; i < len(parts)-1; i++ {
		field := current.FieldByName(parts[i])
		if !field.IsValid() {
			return
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		current = field
	}

	field := current.FieldByName(parts[len(parts)-1])
	if !field.IsValid() {
		return
	}
	field.Set(reflect.ValueOf(value))
}
