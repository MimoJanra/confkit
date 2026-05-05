// Package otel provides OpenTelemetry tracing integration for confkit.
//
// Usage:
//
//	import "github.com/MimoJanra/confkit/otel"
//
//	cfg, err := otel.Load[Config](ctx, tracer,
//	    confkit.FromEnv(),
//	    confkit.FromYAML("config.yaml"),
//	)
package otel

import (
	"context"

	"github.com/MimoJanra/confkit"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Load wraps confkit.Load with an OpenTelemetry span.
func Load[T any](ctx context.Context, tracer trace.Tracer, sources ...confkit.Source) (T, error) {
	ctx, span := tracer.Start(ctx, "confkit.Load")
	defer span.End()

	span.SetAttributes(attribute.Int("confkit.sources", len(sources)))

	cfg, err := confkit.Load[T](sources...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("confkit.success", false))
		return cfg, err
	}

	span.SetAttributes(attribute.Bool("confkit.success", true))
	return cfg, nil
}

// LoadWithOptions wraps confkit.LoadWithOptions with an OpenTelemetry span.
func LoadWithOptions[T any](ctx context.Context, tracer trace.Tracer, options ...confkit.Option) (T, error) {
	ctx, span := tracer.Start(ctx, "confkit.Load")
	defer span.End()

	cfg, err := confkit.LoadWithOptions[T](options...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("confkit.success", false))
		return cfg, err
	}

	span.SetAttributes(attribute.Bool("confkit.success", true))
	return cfg, nil
}
