package confkit

import (
	"context"
	"reflect"
	"time"
)

type Option interface {
	apply(*loadConfig)
}

type AuditEntry struct {
	Field  string
	Source string
	Value  string
}

type AuditLogger func(entries []AuditEntry)

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

func WithAuditLogger(fn AuditLogger) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.AuditLogger = fn
	})
}

func WithLoadHook(fn LoadHookFunc) Option {
	return optionFunc(func(cfg *loadConfig) {
		cfg.LoadHooks = append(cfg.LoadHooks, fn)
	})
}

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
