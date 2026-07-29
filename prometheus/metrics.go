package prometheus

import (
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds the Prometheus collectors confkit exports: confkit_loads_total,
// confkit_load_duration_seconds and confkit_errors_total.
type Metrics struct {
	loadsTotal   *prometheus.CounterVec
	loadDuration prometheus.Histogram
	errorsTotal  *prometheus.CounterVec
}

// NewMetrics creates the collectors and registers them with reg, panicking if any is
// already registered. Pass DefaultRegisterer for the default registry.
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

// DefaultRegisterer is the default Prometheus registry, re-exported so callers need not
// import the client library themselves.
var DefaultRegisterer = prometheus.DefaultRegisterer

// Hook returns a confkit.Option that records each load into these collectors. Pass it
// to confkit.LoadWithOptions.
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
