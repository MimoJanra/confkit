// Package otel provides OpenTelemetry tracing wrappers for confkit.
//
// Import and use in place of confkit.Load:
//
//	import (
//	    "github.com/MimoJanra/confkit/otel"
//	    "go.opentelemetry.io/otel"
//	)
//
//	tracer := otel.Tracer("myapp")
//
//	cfg, err := otelconfkit.Load[Config](ctx, tracer,
//	    confkit.FromYAML("config.yaml"),
//	    confkit.FromEnv(),
//	)
//
// # Tracing
//
// Each Load call creates a "confkit.Load" span. The span includes:
// • confkit.sources — number of sources provided
// • confkit.success — true on success, false on error
//
// When an error occurs, it is recorded on the span and the status is set to Error.
//
// # Context propagation
//
// The provided context (with any active span) is passed into every
// Source.Lookup call via confkit.LoadContext. Cloud sources (Vault, etcd, AWS)
// can use this context to attach child spans or respect deadlines.
//
// # Usage with LoadWithOptions
//
//	cfg, err := otelconfkit.LoadWithOptions[Config](ctx, tracer,
//	    confkit.WithSource(confkit.FromYAML("config.yaml")),
//	    confkit.WithAuditLogger(myLogger),
//	)
package otel
