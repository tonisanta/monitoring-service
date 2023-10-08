package main

import (
	"context"
	"log/slog"
	"monitor-endpoint/internal/annotations"
	"monitor-endpoint/internal/metrics"
	"monitor-endpoint/internal/scheduler"
	"monitor-endpoint/internal/server"
	"monitor-endpoint/internal/service"
	"net/http"
	"time"
)

func main() {
	// TODO: add flags to parse config as input
	slog.Info("Starting monitor service ...")

	m := metrics.NewMetrics()
	ticker := time.NewTicker(time.Second * 10)
	config := service.Config{
		Timeout: time.Second * 2,
	}

	sched := scheduler.NewScheduler(ticker.C)
	myService := service.NewService(m, &http.Client{}, config, time.Now)

	checkUrl := func(ctx context.Context) {
		myService.CheckStatus(ctx, "https://google.com/invented")
	}

	ctx := context.Background()
	go func() {
		sched.Run(ctx, checkUrl)
	}()

	annotationService := annotations.GrafanaAnnotationService{}

	srv := server.NewServer(m, &annotationService)
	srv.Run(ctx)
}
