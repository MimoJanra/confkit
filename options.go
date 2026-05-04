package confkit

import "reflect"

type Option interface {
	apply(*LoadConfig)
}

type LoadConfig struct {
	Sources          []Source
	Validators       map[string]CustomValidatorFunc
	Middleware       []MiddlewareFunc
	InterpolationMax int
}

type MiddlewareFunc func(field FieldInfo, value string) (string, error)

type optionFunc func(*LoadConfig)

func (f optionFunc) apply(cfg *LoadConfig) {
	f(cfg)
}

func WithSource(source Source) Option {
	return optionFunc(func(cfg *LoadConfig) {
		cfg.Sources = append(cfg.Sources, source)
	})
}

func WithValidator(name string, fn func(reflect.Value) error) Option {
	return optionFunc(func(cfg *LoadConfig) {
		cfg.Validators[name] = fn
	})
}

func WithMiddleware(fn MiddlewareFunc) Option {
	return optionFunc(func(cfg *LoadConfig) {
		cfg.Middleware = append(cfg.Middleware, fn)
	})
}

func WithInterpolationMaxDepth(depth int) Option {
	return optionFunc(func(cfg *LoadConfig) {
		if depth > 0 {
			cfg.InterpolationMax = depth
		}
	})
}
