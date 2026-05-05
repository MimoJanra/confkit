package confkit

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func Load[T any](sources ...Source) (T, error) {
	options := make([]Option, 0, len(sources))
	for _, src := range sources {
		options = append(options, WithSource(src))
	}
	return LoadWithOptions[T](options...)
}

func LoadWithWatcher[T any](filePath string, sources ...Source) (T, *ConfigWatcher, error) {
	cfg, err := Load[T](sources...)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	watcher, err := NewConfigWatcher(filePath)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	return cfg, watcher, nil
}

func LoadWithOptions[T any](options ...Option) (T, error) {
	start := time.Now()
	var cfg T

	fields := ScanFields(cfg)
	report := &ErrorReport{}
	parser := NewParser()

	config := &loadConfig{
		Sources:          make([]Source, 0),
		Validators:       make(map[string]CustomValidatorFunc),
		ModelValidators:  make([]func(any) error, 0),
		Middleware:       make([]MiddlewareFunc, 0),
		InterpolationMax: 10,
	}

	for _, opt := range options {
		opt.apply(config)
	}

	validator := NewValidator()
	for name, fn := range config.Validators {
		validator.LocalValidators[name] = fn
	}

	resolver := NewInterpolationResolver(config.InterpolationMax)

	fieldValues := make(map[string]any)
	fieldSources := make(map[string]string)

	for _, field := range fields {
		var sourceErr error
		for _, source := range config.Sources {
			value, ok, err := source.Lookup(&field)
			if err != nil {
				sourceErr = err
				continue
			}
			if ok {
				strVal := anyToString(value)

				middlewareOK := true
				for _, mw := range config.Middleware {
					transformed, err := mw(field, strVal)
					if err != nil {
						report.AddError(FieldError{
							Path:    field.Path,
							Kind:    ErrorKindValidation,
							Message: err.Error(),
						})
						middlewareOK = false
						break
					}
					strVal = transformed
				}

				if !middlewareOK {
					continue
				}

				fieldValues[field.Path] = strVal
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

	interpolationOK := performInterpolation(fieldValues, resolver, report)
	if !interpolationOK {
		fireHooks(config, false, start, len(report.Errors))
		return cfg, report
	}

	val := reflect.ValueOf(&cfg).Elem()
	setStructFields(val, fields, fieldValues, fieldSources, parser, report)

	if !report.IsEmpty() {
		fireHooks(config, false, start, len(report.Errors))
		return cfg, report
	}

	validationErrors := validator.ValidateConfig(cfg, fields)
	if !validationErrors.IsEmpty() {
		report.Errors = append(report.Errors, validationErrors.Errors...)
	}

	// Run model (cross-field) validators only when field validation passed.
	if report.IsEmpty() {
		for _, mv := range config.ModelValidators {
			if err := mv(&cfg); err != nil {
				report.AddError(FieldError{
					Path:    "_model",
					Kind:    ErrorKindValidation,
					Message: err.Error(),
				})
			}
		}
	}

	if !report.IsEmpty() {
		fireHooks(config, false, start, len(report.Errors))
		return cfg, report
	}

	// Audit: record which source each field came from.
	if config.AuditLogger != nil {
		entries := make([]AuditEntry, 0, len(fields))
		for _, field := range fields {
			src, ok := fieldSources[field.Path]
			if !ok {
				continue
			}
			val := anyToString(fieldValues[field.Path])
			if field.IsSecret {
				val = "<redacted>"
			}
			entries = append(entries, AuditEntry{Field: field.Path, Source: src, Value: val})
		}
		config.AuditLogger(entries)
	}

	fireHooks(config, true, start, 0)
	return cfg, nil
}

func fireHooks(config *loadConfig, success bool, start time.Time, errCount int) {
	if len(config.LoadHooks) == 0 {
		return
	}
	d := time.Since(start)
	for _, h := range config.LoadHooks {
		h(success, d, errCount)
	}
}

func performInterpolation(fieldValues map[string]any, resolver *InterpolationResolver, report *ErrorReport) bool {
	for path, rawVal := range fieldValues {
		if rawVal == nil {
			continue
		}
		strVal := anyToString(rawVal)
		resolver.SetConfigValue(path, strVal)
	}

	resolvedValues := make(map[string]string)

	for path, rawVal := range fieldValues {
		if rawVal == nil {
			continue
		}
		strVal := anyToString(rawVal)

		resolved, err := resolver.Resolve(strVal, path)
		if err != nil {
			report.AddError(FieldError{
				Path:    path,
				Kind:    ErrorKindValidation,
				Message: err.Error(),
			})
			return false
		}
		resolvedValues[path] = resolved
	}

	for path, resolved := range resolvedValues {
		fieldValues[path] = resolved
	}

	return true
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
