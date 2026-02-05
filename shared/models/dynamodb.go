package models

import (
	"time"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
)

// DynamoDBEvent represents an event stored in DynamoDB.
type DynamoDBEvent struct {
	Metadata      map[string]any        `json:"metadata"                  dynamodbav:"metadata,omitempty"`
	Amount        Amount                `json:"amount"                    dynamodbav:"amount"`
	EventID       string                `json:"event_id"                  dynamodbav:"PK"`
	AccountID     string                `json:"account_id"                dynamodbav:"SK"`
	Type          constants.EventType   `json:"type"                      dynamodbav:"type"`
	OccurredAt    string                `json:"occurred_at"               dynamodbav:"occurred_at"`
	Source        constants.Source      `json:"source"                    dynamodbav:"source"`
	BatchID       string                `json:"batch_id"                  dynamodbav:"batch_id"`
	Status        constants.EventStatus `json:"status"                    dynamodbav:"status"`
	ReceivedAt    string                `json:"received_at"               dynamodbav:"received_at"`
	ProcessedAt   string                `json:"processed_at,omitempty"    dynamodbav:"processed_at,omitempty"`
	FailedAt      string                `json:"failed_at,omitempty"       dynamodbav:"failed_at,omitempty"`
	LastErrorCode string                `json:"last_error_code,omitempty" dynamodbav:"last_error_code,omitempty"`
	LastErrorMsg  string                `json:"last_error_msg,omitempty"  dynamodbav:"last_error_msg,omitempty"`
	NextRetryAt   string                `json:"next_retry_at,omitempty"   dynamodbav:"next_retry_at,omitempty"`
	StatusDate    string                `json:"status_date,omitempty"     dynamodbav:"status_date,omitempty"` // GSI: status#date.
	Attempts      int                   `json:"attempts"                  dynamodbav:"attempts"`
	TTL           int64                 `json:"ttl,omitempty"             dynamodbav:"ttl,omitempty"`
}

// NewDynamoDBEvent creates a new DynamoDBEvent from a request event.
func NewDynamoDBEvent(event *Event, source constants.Source, batchID string) *DynamoDBEvent {
	now := time.Now().UTC().Format(time.RFC3339)

	return &DynamoDBEvent{
		EventID:       event.EventID,
		AccountID:     event.AccountID,
		Type:          constants.EventType(event.Type),
		OccurredAt:    event.OccurredAt,
		Amount:        event.Amount,
		Metadata:      event.Metadata,
		Source:        source,
		BatchID:       batchID,
		Status:        constants.StatusReceived,
		ReceivedAt:    now,
		Attempts:      0,
		StatusDate:    string(constants.StatusReceived) + "#" + now[:10], // RECEIVED#2025-01-15.
		ProcessedAt:   "",
		FailedAt:      "",
		LastErrorCode: "",
		LastErrorMsg:  "",
		NextRetryAt:   "",
		TTL:           0,
	}
}

// MarkProcessed updates the event status to PROCESSED.
func (e *DynamoDBEvent) MarkProcessed() {
	now := time.Now().UTC().Format(time.RFC3339)
	e.Status = constants.StatusProcessed
	e.ProcessedAt = now
	e.StatusDate = string(constants.StatusProcessed) + "#" + now[:10]
}

// MarkFailed updates the event status to FAILED.
func (e *DynamoDBEvent) MarkFailed(errorCode constants.ErrorCode, errorMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	e.Status = constants.StatusFailed
	e.FailedAt = now
	e.LastErrorCode = string(errorCode)
	e.LastErrorMsg = errorMsg
	e.StatusDate = string(constants.StatusFailed) + "#" + now[:10]
}

// MarkRetried updates the event status to RETRIED.
func (e *DynamoDBEvent) MarkRetried(
	errorCode constants.ErrorCode,
	errorMsg string,
	nextRetry time.Time,
) {
	e.Status = constants.StatusRetried
	e.Attempts++
	e.LastErrorCode = string(errorCode)
	e.LastErrorMsg = errorMsg
	e.NextRetryAt = nextRetry.UTC().Format(time.RFC3339)
}

// CanRetry returns true if the event can be retried.
func (e *DynamoDBEvent) CanRetry() bool {
	return e.Attempts < constants.MaxRetryAttempts && !e.Status.IsTerminal()
}
