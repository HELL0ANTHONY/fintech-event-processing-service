// Package normalization provides interfaces and implementations for data normalization.
package normalization

import (
	"github.com/google/uuid"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// Normalizable defines the contract for normalizing data.
type Normalizable interface {
	Normalize(req *models.Request) *models.Request
}

// Normalizer implements the Normalizable interface.
type Normalizer struct{}

// New creates a new instance of Normalizer.
func New() Normalizable {
	return &Normalizer{}
}

// Normalize processes the input request and ensures required fields are set.
func (n *Normalizer) Normalize(req *models.Request) *models.Request {
	if req.Source == "" {
		req.Source = "api"
	}

	if req.BatchID == "" {
		req.BatchID = uuid.New().String()
	}

	return req
}
