// Package prometheus provides Prometheus metrics integration for confkit.
//
// Usage:
//
//	import "github.com/MimoJanra/confkit/prometheus"
//
//	m := prometheus.NewMetrics(prometheus.DefaultRegisterer)
//	cfg, err := confkit.LoadWithOptions[Config](
//	    confkit.WithSource(confkit.FromEnv()),
//	    m.Hook(),
//	)
package prometheus

import (
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds Prometheus instruments for confkit load operations.
type Metrics struct {
	loadsTotal   *prometheus.CounterVec
	loadDuration prometheus.Histogram
	errorsTotal  *prometheus.CounterVec
}

// NewMetrics creates and registers Prometheus metrics.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		loadsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "confkit_loads_total",
			Help: "Total number of confkit Load calls.",
		}, []string{"status"}),
		loadDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "confkit_load_duration_seconds",
			Help:    "Duration of confkit Load calls in seconds.",
			Buckets: prometheus.DefBuckets,
		}),
		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "confkit_errors_total",
			Help: "Total number of confkit validation errors.",
		}, []string{"kind"}),
	}
	reg.MustRegister(m.loadsTotal, m.loadDuration, m.errorsTotal)
	return m
}

// DefaultRegisterer is prometheus.DefaultRegisterer.
var DefaultRegisterer = prometheus.DefaultRegisterer

// Hook returns a confkit.Option that records metrics for every Load call.
func (m *Metrics) Hook() confkit.Option {
	return confkit.WithLoadHook(func(success bool, d time.Duration, errCount int) {
		if success {
			m.loadsTotal.WithLabelValues("success").Inc()
		} else {
			m.loadsTotal.WithLabelValues("error").Inc()
			m.errorsTotal.WithLabelValues("validation").Add(float64(errCount))
		}
		m.loadDuration.Observe(d.Seconds())
	})
}
