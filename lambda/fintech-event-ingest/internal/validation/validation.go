// Package validation provides interfaces and implementations for validating request schemas and business rules.
package validation

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// Validatable defines the contract for validating request schemas.
type Validatable interface {
	ValidateSchema(data *models.Request) error
	ValidateBusinessRules(data *models.Request) error
}

// Validator implements the Validatable interface.
type Validator struct{}

// New creates a new instance of Validator.
func New() Validatable {
	return &Validator{}
}
