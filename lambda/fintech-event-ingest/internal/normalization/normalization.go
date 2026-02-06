package normalization

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
	"github.com/google/uuid"
)

// Normalizer fills default values for requests.
type Normalizer interface {
	Normalize(req *models.Request) *models.Request
}

type normalizer struct{}

// New creates a new Normalizer.
func New() Normalizer {
	return &normalizer{}
}

// Normalize fills default values for source and batch ID.
func (n *normalizer) Normalize(req *models.Request) *models.Request {
	if req.Source == "" {
		req.Source = "api"
	}

	if req.BatchID == "" {
		req.BatchID = uuid.New().String()
	}

	return req
}
