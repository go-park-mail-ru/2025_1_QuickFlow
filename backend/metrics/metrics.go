package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Hits         *prometheus.CounterVec
	Timings      *prometheus.HistogramVec
	ErrorCounter *prometheus.CounterVec
}

func NewMetrics(namespace string) *Metrics {
	metrics := &Metrics{
		Hits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "Hits",
				Help:      "Total number of HTTP requests",
			},
			[]string{"service", "handler_name", "status"},
		),
		Timings: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "Timings",
				Help:      "HTTP Request duration",
				Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"service", "handler_name"},
		),
		ErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "Errors",
				Help:      "HTTP Errors Number",
			},
			[]string{"service", "handler_name", "status"},
		),
	}

	prometheus.MustRegister(
		metrics.Hits,
		metrics.Timings,
		metrics.ErrorCounter,
	)

	return metrics
}
