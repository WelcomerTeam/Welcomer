package service

import "github.com/prometheus/client_golang/prometheus"

var (
	grpcImgenRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "imgen_grpc_requests_total",
			Help: "Image Generation GRPC Requests",
		},
	)
	grpcImgenTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "imgen_request_total",
			Help: "Image Generation total request count",
		},
		[]string{"guild_id", "format", "background"},
	)

	grpcImgenTotalDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "imgen_request_duration_seconds",
			Help: "Image Generation total request duration",
		},
		[]string{"guild_id", "format", "background"},
	)

	grpcImgenDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "imgen_request_durations_seconds",
			Help:    "Image Generation request durations",
			Buckets: prometheus.ExponentialBucketsRange(0.1, 5, 20),
		},
		[]string{"guild_id", "format", "background"},
	)
)
