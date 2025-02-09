package service

import (
	"time"

	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	imgenRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "imgen_grpc_requests_total",
			Help: "Image Generation Requests",
		},
	)
	imgenTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "imgen_request_total",
			Help: "Image Generation total request count",
		},
		[]string{"guild_id", "format", "background"},
	)

	imgenTotalDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "imgen_request_duration_seconds",
			Help: "Image Generation total request duration",
		},
		[]string{"guild_id", "format", "background"},
	)

	imgenDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "imgen_request_durations_seconds",
			Help:    "Image Generation request durations",
			Buckets: prometheus.ExponentialBucketsRange(0.1, 5, 20),
		},
		[]string{"guild_id", "format", "background"},
	)
)

func onRequest() {
	imgenRequests.Inc()
}

func onGenerationComplete(start time.Time, guildID int64, background string, format utils.ImageFileType) {
	guildIDstring := utils.Itoa(guildID)
	dur := time.Since(start).Seconds()

	imgenTotalRequests.
		WithLabelValues(guildIDstring, format.String(), background).
		Inc()

	imgenTotalDuration.
		WithLabelValues(guildIDstring, format.String(), background).
		Add(dur)

	imgenDuration.
		WithLabelValues(guildIDstring, format.String(), background).
		Observe(dur)
}
