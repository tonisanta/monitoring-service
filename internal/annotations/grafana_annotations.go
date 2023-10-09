package annotations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Config struct {
	Host     string
	ApiToken string
}

type GrafanaAnnotationsRepo struct {
	config     Config
	httpClient *http.Client
}

func NewGrafanaAnnotationsRepo(
	config Config,
	httpClient *http.Client,
) *GrafanaAnnotationsRepo {
	return &GrafanaAnnotationsRepo{
		config:     config,
		httpClient: httpClient,
	}
}

type CreateAnnotationReq struct {
	Time int64    `json:"time"`
	Tags []string `json:"tags"`
	Text string   `json:"text"`
}

func (s *GrafanaAnnotationsRepo) CreateAnnotation(ctx context.Context, annotation Annotation) error {
	slog.Info("creating annotation in grafana ... ")

	var b bytes.Buffer
	createReq := MapToCreateAnnotationReq(annotation)
	err := json.NewEncoder(&b).Encode(createReq)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.config.Host+"/api/annotations", &b)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.config.ApiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	slog.Info("creating annotation in grafana - done ")
	return nil
}

func MapToCreateAnnotationReq(annotation Annotation) CreateAnnotationReq {
	return CreateAnnotationReq{
		Time: annotation.Time.UnixMilli(),
		Tags: annotation.Tags,
		Text: annotation.Text,
	}
}
