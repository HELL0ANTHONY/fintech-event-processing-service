// Package validation provides business rule validation for events.
package validation

import (
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
)

// BusinessValidator validates events against business rules.
type BusinessValidator interface {
	ValidateBusinessRules(event *models.DynamoDBEvent) error
}

type validator struct{}

// New creates a new BusinessValidator.
func New() BusinessValidator {
	return &validator{}
}
