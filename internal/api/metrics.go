package api

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "authz_http_requests_total",
			Help: "Total HTTP request yang diterima, per endpoint dan status code.",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "authz_http_request_duration_seconds",
			Help:    "Distribusi durasi HTTP request.",
			Buckets: prometheus.DefBuckets, // 5ms - 10s
		},
		[]string{"path", "method"},
	)

	decisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "authz_decisions_total",
			Help: "Total keputusan authorization, per kind (can/check_relation) dan hasil.",
		},
		[]string{"kind", "allowed"},
	)

	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "authz_cache_hits_total",
			Help: "Total cache hit vs miss untuk keputusan authorization.",
		},
		[]string{"hit"},
	)

	rateLimitRejectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "authz_rate_limit_rejected_total",
			Help: "Total request yang ditolak karena melebihi rate limit.",
		},
	)
)

func RecordDecision(kind string, allowed, fromCache bool) {
	decisionsTotal.WithLabelValues(kind, strconv.FormatBool(allowed)).Inc()
	cacheHitsTotal.WithLabelValues(strconv.FormatBool(fromCache)).Inc()
}
