package validation

import (
	"regexp"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// Validatable defines the contract for validating request schemas.
type Validatable interface {
	ValidateSchema(data *models.Request) error
}

// Validator implements the Validatable interface.
type Validator struct {
	regexs map[string]*regexp.Regexp
}

// NewValidator creates a new instance of Validator.
func New() Validatable {
	return &Validator{}
}
