// Package constants contains constant definitions used across the fintech event ingest service.
package constants

// EventStatus defines the possible statuses for an event.
type EventStatus string

// MaxEventsPerRequest defines the maximum number of events allowed in a single request.
var MaxEventsPerRequest = 500

// Predefined event statuses.
const (
	StatusFailed    EventStatus = "FAILED"
	StatusProcessed EventStatus = "PROCESSED"
	StatusReceived  EventStatus = "RECEIVED"
	StatusRetried   EventStatus = "RETRIED"
)

// EventType defines the types of events that can be processed.
type EventType string

// Predefined event types.
const (
	EventMovementPosted   EventType = "movement_posted"
	EventMovementReversed EventType = "movement_reversed"
	EventMovementAdjusted EventType = "movement_adjusted"

	EventCreditSettled  EventType = "credit_settled"
	EventCreditReversed EventType = "credit_reversed"

	EventReconciliationResult EventType = "reconciliation_result"

	EventBatchReceived  EventType = "batch_received"
	EventBatchProcessed EventType = "batch_processed"
	EventBatchFailed    EventType = "batch_failed"
)

// IsSupportedEventType checks if the given event type is supported.
func IsSupportedEventType(et EventType) bool {
	switch et {
	case EventMovementPosted,
		EventMovementReversed,
		EventMovementAdjusted,
		EventCreditSettled,
		EventCreditReversed,
		EventReconciliationResult,
		EventBatchReceived,
		EventBatchProcessed,
		EventBatchFailed:
		return true
	default:
		return false
	}
}
