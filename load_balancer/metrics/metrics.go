package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// общие HTTP-метрики
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests seen by the limiter (label status shows final HTTP code).",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests passed through the limiter.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active connections handled by the limiter.",
		},
	)

	// метрики прокси-балансера
	ProxiedRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxied_requests_total",
			Help: "Total number of requests proxied to each backend.",
		},
		[]string{"backend"},
	)

	ProxiedFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxied_failures_total",
			Help: "Number of failed proxy attempts per backend.",
		},
		[]string{"backend"},
	)

	BackendResponseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "backend_response_status",
			Help: "HTTP status codes returned from each backend.",
		},
		[]string{"backend", "status"},
	)

	BackendConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "backend_active_connections",
			Help: "Current number of active connections per backend.",
		},
		[]string{"backend"},
	)
)

func Init() {
	prometheus.MustRegister(
		RequestCount,
		RequestDuration,
		ActiveConnections,
		ProxiedRequestCount,
		ProxiedFailuresTotal,
		BackendResponseStatus,
		BackendConnections,
	)
}
