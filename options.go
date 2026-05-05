package confkit

import "reflect"

type Option interface {
	apply(*loadConfig)
}

type loadConfig struct {
	Sources          []Source
	Validators       map[string]CustomValidatorFunc
	Middleware       []MiddlewareFunc
	InterpolationMax int
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
