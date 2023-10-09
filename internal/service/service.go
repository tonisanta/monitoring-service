package service

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"monitor-endpoint/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	Timeout time.Duration
}

type TimeProvider func() time.Time

type Service struct {
	metrics      *metrics.Metrics
	httpClient   *http.Client
	config       Config
	timeProvider TimeProvider
}

func NewService(
	metrics *metrics.Metrics,
	httpClient *http.Client,
	config Config,
	timeProvider TimeProvider,
) *Service {
	return &Service{
		metrics:      metrics,
		httpClient:   httpClient,
		config:       config,
		timeProvider: timeProvider,
	}
}

func (s *Service) CheckStatus(parentCtx context.Context, url string) {
	ctx, cancel := context.WithTimeout(parentCtx, s.config.Timeout)
	defer cancel()

	lab := prometheus.Labels{
		metrics.LabelEndpoint: url,
	}
	s.metrics.NumActiveRequests.With(lab).Inc()
	defer func() {
		s.metrics.NumActiveRequests.With(lab).Dec()
	}()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	req = req.WithContext(ctx)

	start := s.timeProvider()
	response, err := s.httpClient.Do(req)
	defer func() {
		if response != nil {
			err := response.Body.Close()
			if err != nil {
				slog.Error("failed to close response", slog.String("error", err.Error()))
			}
		}
	}()

	var statusCode *int
	if err != nil {
		slog.Error(err.Error())
		if errors.Is(err, context.DeadlineExceeded) {
			statusCode = Ptr(499)
			slog.Error("timeout, client closed request")
		}
	}

	if response != nil && statusCode == nil {
		statusCode = &response.StatusCode
	}

	labels := prometheus.Labels{
		metrics.LabelEndpoint: url,
	}

	if statusCode != nil {
		labels[metrics.LabelCode] = strconv.Itoa(*statusCode)
	} else {
		labels[metrics.LabelCode] = "no status code"
	}

	elapsedTime := s.timeProvider().Sub(start)
	s.metrics.ResponseTime.With(labels).Observe(elapsedTime.Seconds())
	s.metrics.NumRequests.With(labels).Inc()
}

func Ptr[T any](o T) *T {
	return &o
}
