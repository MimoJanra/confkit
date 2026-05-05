package confkit

import (
	"reflect"
	"time"
)

type Option interface {
	apply(*loadConfig)
}

// AuditEntry records a single field's resolved value and its source.
type AuditEntry struct {
	Field  string
	Source string
	Value  string // redacted when field is secret
}

// AuditLogger is called after a successful load with the full list of resolved fields.
type AuditLogger func(entries []AuditEntry)

// LoadHookFunc is called after every load attempt with the outcome.
type LoadHookFunc func(success bool, duration time.Duration, errCount int)

type loadConfig struct {
	Sources          []Source
	Validators       map[string]CustomValidatorFunc
	ModelValidators  []func(any) error
	Middleware       []MiddlewareFunc
	InterpolationMax int
	AuditLogger      AuditLogger
	LoadHooks        []LoadHookFunc
}

type MiddlewareFunc func(field FieldInfo, value string) (string, error)

type optionFunc func(*loadConfig)

func (f optionFunc) apply(cfg *loadConfig) {
	f(cfg)
}

func WithSource(source Source) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Sources = append(cfg.Sources, source)
	})
}

func WithValidator(name string, fn func(reflect.Value) error) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Validators[name] = fn
	})
}

// WithModelValidator registers a cross-field validator that runs after all field validators pass.
// The function receives a pointer to the fully populated config struct.
func WithModelValidator[T any](fn func(*T) error) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.ModelValidators = append(cfg.ModelValidators, func(v any) error {
			if typed, ok := v.(*T); ok {
				return fn(typed)
			}
			return nil
		})
	})
}

func WithMiddleware(fn MiddlewareFunc) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Middleware = append(cfg.Middleware, fn)
	})
}

func WithInterpolationMaxDepth(depth int) Option {
	return optionFunc(func(cfg *loadConfig) {
		if depth > 0 {
			cfg.InterpolationMax = depth
		}
	})
}

// WithAuditLogger registers a callback that receives every resolved field after a successful load.
func WithAuditLogger(fn AuditLogger) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.AuditLogger = fn
	})
}

// WithLoadHook registers a callback invoked after every load with success status and duration.
// Useful for metrics collection (see confkit/prometheus and confkit/otel submodules).
func WithLoadHook(fn LoadHookFunc) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.LoadHooks = append(cfg.LoadHooks, fn)
	})
}
