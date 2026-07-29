package confkit

import (
	"context"
	"reflect"
	"time"
)

// Option configures a load performed by LoadWithOptions. The interface is
// deliberately closed: options come from the With* constructors in this package.
type Option interface {
	apply(*loadConfig)
}

// AuditEntry records where one field's value came from. Values of fields tagged
// `secret:"true"` are redacted before the entry is created.
type AuditEntry struct {
	Field  string
	Source string
	Value  string
}

// AuditLogger receives one entry per resolved field after a load. Register it with
// WithAuditLogger. ValidateOnly deliberately does not invoke it.
type AuditLogger func(entries []AuditEntry)

// LoadHookFunc is called once a load finishes, with whether it succeeded, how long
// it took and how many errors were reported. Register it with WithLoadHook; it
// backs the Prometheus and OpenTelemetry integrations.
type LoadHookFunc func(success bool, duration time.Duration, errCount int)

type loadConfig struct {
	Sources          []Source
	Validators       map[string]CustomValidatorFunc
	ModelValidators  []func(any) error
	Middleware       []MiddlewareFunc
	InterpolationMax int
	AuditLogger      AuditLogger
	LoadHooks        []LoadHookFunc
	Ctx              context.Context
	validateOnlyMode bool
}

// MiddlewareFunc transforms a raw value after a Source produced it but before it is
// parsed, which makes it suitable for trimming, decoding or decrypting. Returning
// an error fails that field. Register it with WithMiddleware.
type MiddlewareFunc func(field FieldInfo, value string) (string, error)

type optionFunc func(*loadConfig)

func (f optionFunc) apply(cfg *loadConfig) {
	f(cfg)
}

// WithSource appends a Source. Order matters: later sources override earlier ones.
func WithSource(source Source) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Sources = append(cfg.Sources, source)
	})
}

// WithValidator registers a named rule usable from a `validate` tag, so
// WithValidator("even", fn) enables `validate:"even"`. A name matching a built-in
// rule shadows it for this load only.
func WithValidator(name string, fn func(reflect.Value) error) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Validators[name] = fn
	})
}

// WithModelValidator registers a whole-config check, for rules that span several
// fields such as "TLS requires both a cert and a key". It runs only once every
// per-field rule has passed, and its error is reported against the path "_model".
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

// WithMiddleware appends a value transformer. Middleware runs in registration order,
// each receiving the previous one's output.
func WithMiddleware(fn MiddlewareFunc) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Middleware = append(cfg.Middleware, fn)
	})
}

// WithInterpolationMaxDepth caps how deeply ${VAR} references may nest before the
// load fails, which is what stops a reference cycle from recursing forever. The
// default is 10; values of zero or less are ignored.
func WithInterpolationMaxDepth(depth int) Option {
	return optionFunc(func(cfg *loadConfig) {
		if depth > 0 {
			cfg.InterpolationMax = depth
		}
	})
}

// WithAuditLogger sets the audit logger, replacing any previous one. It is called
// on both success and failure, but never by ValidateOnly.
func WithAuditLogger(fn AuditLogger) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.AuditLogger = fn
	})
}

// WithLoadHook appends a completion hook. Hooks run in registration order and are
// skipped by ValidateOnly.
func WithLoadHook(fn LoadHookFunc) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.LoadHooks = append(cfg.LoadHooks, fn)
	})
}

// WithContext sets the context handed to every Source. It overrides the context
// argument of LoadWithOptionsContext.
func WithContext(ctx context.Context) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.Ctx = ctx
	})
}

func withValidateOnlyMode() Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.validateOnlyMode = true
	})
}
