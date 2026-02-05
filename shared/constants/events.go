package constants

// MaxEventsPerRequest defines the maximum number of events allowed in a single request.
const MaxEventsPerRequest = 500

// EventType defines the types of events that can be processed.
type EventType string

const (
	// Movement events.
	EventMovementPosted   EventType = "movement_posted"
	EventMovementReversed EventType = "movement_reversed"
	EventMovementAdjusted EventType = "movement_adjusted"

	// Credit events.
	EventCreditSettled  EventType = "credit_settled"
	EventCreditReversed EventType = "credit_reversed"

	// Reconciliation events.
	EventReconciliationResult EventType = "reconciliation_result"

	// Batch events.
	EventBatchReceived  EventType = "batch_received"
	EventBatchProcessed EventType = "batch_processed"
	EventBatchFailed    EventType = "batch_failed"
)

// IsSupportedEventType checks if the given event type is supported.
func IsSupportedEventType(et EventType) bool {
	switch et {
	case EventMovementPosted, EventMovementReversed, EventMovementAdjusted,
		EventCreditSettled, EventCreditReversed,
		EventReconciliationResult,
		EventBatchReceived, EventBatchProcessed, EventBatchFailed:
		return true
	default:
		return false
	}
}

// RequiresAmount returns true if the event type requires an amount.
func (et EventType) RequiresAmount() bool {
	switch et {
	case EventMovementPosted, EventMovementReversed, EventMovementAdjusted,
		EventCreditSettled, EventCreditReversed:
		return true
	default:
		return false
	}
}
