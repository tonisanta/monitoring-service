package annotations

import (
	"context"
	"log"
)

type GrafanaAnnotationService struct {
}

func (s *GrafanaAnnotationService) CreateAnnotation(ctx context.Context, annotation Annotation) error {
	log.Println("creation annotation in grafana")
	log.Printf("created: %+v", annotation)
	// TODO: Create real implementation
	return nil
}
