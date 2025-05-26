package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Register() {
	prometheus.MustRegister(HTTPRequestCount)
	prometheus.MustRegister(HTTPRequestDuration)
}
