package otel

import (
	"context"

	"github.com/MimoJanra/confkit"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Load wraps confkit.LoadContext with an OpenTelemetry span.
// The span context is propagated into source lookups so cloud sources
// (Vault, etcd, AWS) can attach their own spans as children.
func Load[T any](ctx context.Context, tracer trace.Tracer, sources ...confkit.Source) (*T, error) {
	ctx, span := tracer.Start(ctx, "confkit.Load")
	defer span.End()

	span.SetAttributes(attribute.Int("confkit.sources", len(sources)))

	cfg, err := confkit.LoadContext[T](ctx, sources...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("confkit.success", false))
		return cfg, err
	}

	span.SetAttributes(attribute.Bool("confkit.success", true))
	return cfg, nil
}

// LoadWithOptions wraps confkit.LoadWithOptionsContext with an OpenTelemetry span.
// The span context is propagated into source lookups via WithContext.
func LoadWithOptions[T any](ctx context.Context, tracer trace.Tracer, options ...confkit.Option) (*T, error) {
	ctx, span := tracer.Start(ctx, "confkit.Load")
	defer span.End()

	cfg, err := confkit.LoadWithOptionsContext[T](ctx, options...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("confkit.success", false))
		return cfg, err
	}

	span.SetAttributes(attribute.Bool("confkit.success", true))
	return cfg, nil
}
