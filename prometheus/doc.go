// Package prometheus provides Prometheus metrics integration for confkit.
//
// Use Metrics.Hook() to attach a load hook that records:
// • confkit_loads_total{status="success|error"} — counter
// • confkit_load_duration_seconds — histogram
// • confkit_errors_total{kind="validation"} — counter
//
// # Usage
//
//	import (
//	    "github.com/MimoJanra/confkit"
//	    ckprom "github.com/MimoJanra/confkit/prometheus"
//	    "github.com/prometheus/client_golang/prometheus"
//	)
//
//	metrics := ckprom.NewMetrics(prometheus.DefaultRegisterer)
//
//	cfg, err := confkit.LoadWithOptions[Config](
//	    confkit.WithSource(confkit.FromYAML("config.yaml")),
//	    confkit.WithSource(confkit.FromEnv()),
//	    metrics.Hook(),
//	)
//
// # Custom registry
//
// Pass any prometheus.Registerer to NewMetrics:
//
//	reg := prometheus.NewRegistry()
//	metrics := ckprom.NewMetrics(reg)
//
// # Metrics reference
//
//	confkit_loads_total          — total Load calls, labelled by status
//	confkit_load_duration_seconds — Load latency histogram
//	confkit_errors_total          — validation errors per Load, labelled by kind
package prometheus
