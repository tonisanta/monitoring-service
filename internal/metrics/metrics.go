package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	LabelCode     = "code"
	LabelEndpoint = "endpoint"
)

type Metrics struct {
	Reg               *prometheus.Registry
	Factory           promauto.Factory
	NumRequests       *prometheus.CounterVec
	ResponseTime      *prometheus.HistogramVec
	NumActiveRequests *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	factory := promauto.With(reg)

	numRequests := factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "num_requests",
			Help: "Total number of HTTP requests by status code and endpoint.",
		}, []string{LabelCode, LabelEndpoint},
	)

	numActiveRequests := factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "num_active_requests",
			Help: "Total number of active HTTP requests by endpoint.",
		}, []string{LabelEndpoint},
	)

	responseTime := factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_time",
			Help:    "A histogram based on the response time by status code and endpoint.",
			Buckets: prometheus.DefBuckets,
		}, []string{LabelCode, LabelEndpoint},
	)

	return &Metrics{
		Reg:               reg,
		Factory:           factory,
		NumRequests:       numRequests,
		ResponseTime:      responseTime,
		NumActiveRequests: numActiveRequests,
	}
}
