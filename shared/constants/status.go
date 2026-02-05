package constants

// EventStatus defines the possible statuses for an event in the processing pipeline.
type EventStatus string

const (
	StatusFailed    EventStatus = "FAILED"
	StatusProcessed EventStatus = "PROCESSED"
	StatusReceived  EventStatus = "RECEIVED"
	StatusRetried   EventStatus = "RETRIED"
)

// IsTerminal returns true if the status is a terminal state.
func (s EventStatus) IsTerminal() bool {
	return s == StatusProcessed || s == StatusFailed
}
