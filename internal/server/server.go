package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"log/slog"
	"monitor-endpoint/internal/annotations"
	"monitor-endpoint/internal/metrics"
	"net/http"
	"os/exec"
	"time"
)

const (
	Port = 8080
)

type AnnotationsService interface {
	CreateAnnotation(context.Context, annotations.Annotation) error
}

type Server struct {
	metrics           *metrics.Metrics
	annotationService AnnotationsService
}

func NewServer(
	metrics *metrics.Metrics,
	annotationService AnnotationsService,
) *Server {
	return &Server{
		metrics:           metrics,
		annotationService: annotationService,
	}
}

func (s *Server) Run(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(s.metrics.Reg, promhttp.HandlerOpts{Registry: s.metrics.Reg}))
	mux.Handle("/exec", http.HandlerFunc(s.execHandler))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Port),
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

	cmd := exec.Command(req.Command)
	if errors.Is(cmd.Err, exec.ErrDot) {
		log.Println("inside here")
		cmd.Err = nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	annotation := annotations.Annotation{
		Text: "My custom text",
		Time: time.Now(),
		Tags: []string{"tag1"},
	}

	err = s.annotationService.CreateAnnotation(r.Context(), annotation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}
