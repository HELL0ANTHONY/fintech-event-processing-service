// Package repository provides data access abstractions for the event processing service.
package repository

import (
	"context"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/models"
)

// EventRepository defines the contract for event persistence operations.
type EventRepository interface {
	// SaveEvent persists a single event with conditional check for idempotency.
	// Returns ErrDuplicateEvent if the event already exists.
	SaveEvent(ctx context.Context, event *models.DynamoDBEvent) error

	// SaveEventsBatch persists multiple events in a batch.
	// Returns a slice of results indicating success/failure for each event.
	SaveEventsBatch(ctx context.Context, events []*models.DynamoDBEvent) ([]BatchResult, error)

	// UpdateEventStatus updates the status and related fields of an event.
	UpdateEventStatus(ctx context.Context, event *models.DynamoDBEvent) error

	// GetEvent retrieves an event by its ID.
	GetEvent(ctx context.Context, eventID, accountID string) (*models.DynamoDBEvent, error)

	// QueryEventsByStatus retrieves events by status and date range.
	QueryEventsByStatus(
		ctx context.Context,
		status string,
		date string,
		limit int32,
	) ([]*models.DynamoDBEvent, error)
}

// BatchResult represents the result of a batch operation for a single item.
type BatchResult struct {
	Error     error
	EventID   string
	Success   bool
	Duplicate bool
}
