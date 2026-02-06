package validation

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
)

// SchemaValidator validates events against the expected schema.
type SchemaValidator interface {
	ValidateSchema(event *models.Event) error
}

type validator struct{}

// New creates a new SchemaValidator.
func New() SchemaValidator {
	return &validator{}
}
