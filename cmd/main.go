package main

import (
	"context"
	"crypto/tls"
	"flag"
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
	slog.Info("Starting monitor service ...")

	urlFlag := flag.String("url", "https://google.com/", "Url to be monitored")
	apiTokenFlag := flag.String("token", "insert your token", "Grafana API token")
	grafanaHost := flag.String("host", "http://localhost:3000", "Grafana host")
	flag.Parse()

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
		myService.CheckStatus(ctx, *urlFlag)
	}

	ctx := context.Background()
	go func() {
		sched.Run(ctx, checkUrl)
	}()

	grafanaConfig := annotations.Config{
		Host:     *grafanaHost,
		ApiToken: *apiTokenFlag,
	}
	annotationsRepo := annotations.NewGrafanaAnnotationsRepo(grafanaConfig, &http.Client{})
	srv := server.NewServer(m, annotationsRepo)
	srv.Run(ctx)
}
