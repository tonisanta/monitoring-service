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

	urlFlag := flag.String("url", "https://google.com/", "Url to monitor")
	apiTokenFlag := flag.String("token", "insert your token", "Grafana API token")
	grafanaHost := flag.String("host", "http://localhost:3000", "Grafana host")
	tickerFreqFlag := flag.String("frequency", "1m", `Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)
	timeoutFlag := flag.String("timeout", "30s", `Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)
	portFlag := flag.Int("port", 8080, "Port used by the server")
	flag.Parse()

	tickerFreq, err := time.ParseDuration(*tickerFreqFlag)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	timeout, err := time.ParseDuration(*timeoutFlag)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	m := metrics.NewMetrics()
	ticker := time.NewTicker(tickerFreq)
	config := service.Config{
		Timeout: timeout,
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
	srvConfig := server.Config{
		Port: *portFlag,
	}
	srv := server.NewServer(m, srvConfig, annotationsRepo)
	srv.Run(ctx)
}
