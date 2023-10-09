package main

import (
	"context"
	"crypto/tls"
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
	ticker := time.NewTicker(time.Minute * 1)
	config := service.Config{
		Timeout: time.Second * 30,
	}

	sched := scheduler.NewScheduler(ticker.C)
	myService := service.NewService(m, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}, config, time.Now)

	checkUrl := func(ctx context.Context) {
		myService.CheckStatus(ctx, "https://google.com/")
	}

	ctx := context.Background()
	go func() {
		sched.Run(ctx, checkUrl)
	}()

	grafanaConfig := annotations.Config{
		Host:     "http://localhost:3000",
		ApiToken: "",
	}
	annotationsRepo := annotations.NewGrafanaAnnotationsRepo(grafanaConfig, &http.Client{})
	srv := server.NewServer(m, annotationsRepo)
	srv.Run(ctx)
}
