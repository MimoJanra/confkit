package confkit

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/MimoJanra/confkit/internal/interpolation"
	"github.com/MimoJanra/confkit/internal/parser"
)

// Load populates a new T from the given sources and validates the result.
//
// Sources are consulted in order and the first one holding a field wins, so pass
// them from highest to lowest precedence — typically flags, then environment, then
// file. Fields no source provides fall back to their `default` tag.
//
// On failure the error is an *ErrorReport describing every problem found rather
// than just the first; use Explain to format it for humans. The returned pointer
// is non-nil even on failure, but only partially populated.
func Load[T any](sources ...Source) (*T, error) {
	return LoadContext[T](context.Background(), sources...)
}

// LoadContext is Load with a context, which is passed to each Source. Use it when
// a source performs I/O, such as the Vault, Consul, etcd and AWS sources.
func LoadContext[T any](ctx context.Context, sources ...Source) (*T, error) {
	opts := make([]Option, 0, len(sources)+1)
	opts = append(opts, WithContext(ctx))
	for _, src := range sources {
		opts = append(opts, WithSource(src))
	}
	return loadCore[T](opts...)
}

// LoadWithOptions populates a new T using functional options, which extend Load
// with custom validators, middleware, audit logging and load hooks. Sources are
// supplied through WithSource.
func LoadWithOptions[T any](options ...Option) (*T, error) {
	return LoadWithOptionsContext[T](context.Background(), options...)
}

// LoadWithOptionsContext is LoadWithOptions with a context. An explicit
// WithContext option, if given, takes precedence over ctx.
func LoadWithOptionsContext[T any](ctx context.Context, options ...Option) (*T, error) {
	opts := make([]Option, 0, len(options)+1)
	opts = append(opts, WithContext(ctx))
	opts = append(opts, options...)
	return loadCore[T](opts...)
}

// ValidateOnly runs the full load pipeline (sources + validation) but skips LoadHookFunc and AuditLogger.
// Use in CI to validate config without triggering side effects.
func ValidateOnly[T any](ctx context.Context, options ...Option) (*T, error) {
	opts := make([]Option, 0, len(options)+2)
	opts = append(opts, WithContext(ctx), withValidateOnlyMode())
	opts = append(opts, options...)
	return loadCore[T](opts...)
}

// MustLoad is Load but panics on failure. Intended for package-level variables
// and program start-up, where a bad configuration should stop the process.
func MustLoad[T any](sources ...Source) *T {
	cfg, err := Load[T](sources...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// MustLoadContext is LoadContext but panics on failure.
func MustLoadContext[T any](ctx context.Context, sources ...Source) *T {
	cfg, err := LoadContext[T](ctx, sources...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// LoadWithWatcher loads a config from sources and also returns a ConfigWatcher for
// filePath.
//
// The watcher is returned stopped: call Start to begin polling, and Stop when
// finished. Stop is safe even if Start was never called.
func LoadWithWatcher[T any](filePath string, sources ...Source) (*T, *ConfigWatcher, error) {
	cfg, err := Load[T](sources...)
	if err != nil {
		return nil, nil, err
	}

	watcher, err := NewConfigWatcher(filePath)
	if err != nil {
		return nil, nil, err
	}

	return cfg, watcher, nil
}

func loadCore[T any](options ...Option) (*T, error) {
	start := time.Now()
	cfg := new(T)

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

	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	fields := ScanFields(cfg)
	report := &ErrorReport{}
	prs := parser.New()

	validator := NewValidator()
	for name, fn := range config.Validators {
		validator.LocalValidators[name] = fn
	}

	resolver := interpolation.NewResolver(config.InterpolationMax)

	fieldValues := make(map[string]any)
	fieldSources := make(map[string]string)

	for _, field := range fields {
		var sourceErr error
		for _, source := range config.Sources {
			value, ok, err := source.Lookup(ctx, &field)
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
					break
				}

				fieldValues[field.Path] = strVal
				fieldSources[field.Path] = source.Name()
				break
			}
		}
		if _, found := fieldValues[field.Path]; !found && sourceErr != nil {
			report.AddError(FieldError{
				Path:    field.Path,
				Kind:    ErrorKindIO,
				Message: sourceErr.Error(),
				Err:     sourceErr,
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
		if !config.validateOnlyMode {
			callAuditLogger(config, fieldValues, fieldSources, fields)
			fireHooks(config, false, start, len(report.Errors))
		}
		return cfg, report
	}

	val := reflect.ValueOf(cfg).Elem()
	initEmbeddedPointers(val, val.Type())
	setStructFields(val, fields, fieldValues, fieldSources, prs, report)

	if !report.IsEmpty() {
		if !config.validateOnlyMode {
			callAuditLogger(config, fieldValues, fieldSources, fields)
			fireHooks(config, false, start, len(report.Errors))
		}
		return cfg, report
	}

	validationErrors := validator.ValidateConfig(cfg, fields)
	if !validationErrors.IsEmpty() {
		report.Errors = append(report.Errors, validationErrors.Errors...)
	}

	if report.IsEmpty() {
		for _, mv := range config.ModelValidators {
			if err := mv(cfg); err != nil {
				report.AddError(FieldError{
					Path:    "_model",
					Kind:    ErrorKindValidation,
					Message: err.Error(),
				})
			}
		}
	}

	if !report.IsEmpty() {
		if !config.validateOnlyMode {
			callAuditLogger(config, fieldValues, fieldSources, fields)
			fireHooks(config, false, start, len(report.Errors))
		}
		return cfg, report
	}

	if !config.validateOnlyMode {
		callAuditLogger(config, fieldValues, fieldSources, fields)
		fireHooks(config, true, start, 0)
	}
	return cfg, nil
}

func callAuditLogger(config *loadConfig, fieldValues map[string]any, fieldSources map[string]string, fields []FieldInfo) {
	if config.AuditLogger == nil {
		return
	}
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

func fireHooks(config *loadConfig, success bool, start time.Time, errCount int) {
	if len(config.LoadHooks) == 0 {
		return
	}
	d := time.Since(start)
	for _, h := range config.LoadHooks {
		h(success, d, errCount)
	}
}

func performInterpolation(fieldValues map[string]any, resolver *interpolation.Resolver, report *ErrorReport) bool {
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

func setStructFields(val reflect.Value, fields []FieldInfo, values map[string]any, sources map[string]string, prs *parser.Parser, report *ErrorReport) {
	for _, field := range fields {
		rawVal, ok := values[field.Path]
		if !ok {
			continue
		}

		strVal := anyToString(rawVal)
		parsed, err := prs.Parse(strVal, field.Type)
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
		if field.Kind() == reflect.Pointer {
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

	v := reflect.ValueOf(value)
	// The parser returns the builtin underlying type (e.g. string for a field
	// declared as `type Level string`), which reflect.Set would reject outright.
	// Convert when the kinds already match, which the parser guarantees.
	if ft := field.Type(); v.Type() != ft && v.Kind() == ft.Kind() && v.Type().ConvertibleTo(ft) {
		v = v.Convert(ft)
	}
	field.Set(v)
}
