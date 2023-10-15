package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"monitor-endpoint/internal/annotations"
	"monitor-endpoint/internal/metrics"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	Port = 8080
)

type AnnotationsRepository interface {
	CreateAnnotation(context.Context, annotations.Annotation) error
}

type Config struct {
	Port int
}

type Server struct {
	metrics              *metrics.Metrics
	config               Config
	annotationRepository AnnotationsRepository
}

func NewServer(
	metrics *metrics.Metrics,
	config Config,
	annotationService AnnotationsRepository,
) *Server {
	return &Server{
		metrics:              metrics,
		config:               config,
		annotationRepository: annotationService,
	}
}

func (s *Server) Run(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(s.metrics.Reg, promhttp.HandlerOpts{Registry: s.metrics.Reg}))
	mux.Handle("/exec", http.HandlerFunc(s.execHandler))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error(err.Error(), slog.String("server", "exposing metrics"))
	}
}

type ExecRequest struct {
	Command string `json:"command"`
}

func (s *Server) execHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	args := strings.Fields(req.Command)
	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	annotation := annotations.Annotation{
		Text: "Exec: " + req.Command,
		Time: time.Now(),
		Tags: []string{"command"},
	}

	err = s.annotationRepository.CreateAnnotation(r.Context(), annotation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}
