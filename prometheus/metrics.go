package prometheus

import (
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	loadsTotal   *prometheus.CounterVec
	loadDuration prometheus.Histogram
	errorsTotal  *prometheus.CounterVec
}

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

var DefaultRegisterer = prometheus.DefaultRegisterer

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
