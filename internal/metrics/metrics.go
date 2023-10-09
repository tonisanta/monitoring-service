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
	Reg                 *prometheus.Registry
	Factory             promauto.Factory
	NumRequests         *prometheus.CounterVec
	ResponseTime        *prometheus.HistogramVec
	NumActiveRequests   *prometheus.GaugeVec
	ResponseTimeSummary *prometheus.SummaryVec
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

	responseTimeSummary := factory.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "summary_response_time_seconds",
			Help:       "A summary based on the response time by status code and endpoint.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}, []string{LabelCode, LabelEndpoint},
	)

	return &Metrics{
		Reg:                 reg,
		Factory:             factory,
		NumRequests:         numRequests,
		ResponseTime:        responseTime,
		NumActiveRequests:   numActiveRequests,
		ResponseTimeSummary: responseTimeSummary,
	}
}
