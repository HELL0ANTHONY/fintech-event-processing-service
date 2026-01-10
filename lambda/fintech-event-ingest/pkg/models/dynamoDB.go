package models

import "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"

// DynamoDBEvent represents the structure of an event stored in DynamoDB.
type DynamoDBEvent struct {
	Metadata      map[string]any        `json:"metadata"                  dynamodbav:"metadata"`
	Amount        Amount                `json:"amount"                    dynamodbav:"amount"`
	Status        constants.EventStatus `json:"status"                    dynamodbav:"status"`
	ReceivedAt    string                `json:"received_at"               dynamodbav:"received_at"`
	ProcessedAt   string                `json:"processed_at,omitempty"    dynamodbav:"processed_at,omitempty"`
	FailedAt      string                `json:"failed_at,omitempty"       dynamodbav:"failed_at,omitempty"`
	LastErrorCode string                `json:"last_error_code,omitempty" dynamodbav:"last_error_code,omitempty"`
	NextRetryAt   string                `json:"next_retry_at,omitempty"   dynamodbav:"next_retry_at,omitempty"`
	BatchID       string                `json:"batch_id"                  dynamodbav:"batch_id"`
	EventID       string                `json:"event_id"                  dynamodbav:"event_id"`
	Type          constants.EventType   `json:"type"                      dynamodbav:"type"`
	OccurredAt    string                `json:"occurred_at"               dynamodbav:"occurred_at"`
	AccountID     string                `json:"account_id"                dynamodbav:"account_id"`
	Source        constants.Source      `json:"source"                    dynamodbav:"source"`
	Attempts      int                   `json:"attempts"                  dynamodbav:"attempts"`
}

// NewDynamoDBEventFromRequest creates a new DynamoDBEvent from a Request.
func NewDynamoDBEventFromRequest(
	event *Event,
	source constants.Source,
	receivedAt, batchID string,
) *DynamoDBEvent {
	return &DynamoDBEvent{
		Source:        source,
		BatchID:       batchID,
		ReceivedAt:    receivedAt,
		Status:        constants.StatusReceived,
		Attempts:      0,
		Metadata:      event.Metadata,
		Amount:        event.Amount,
		EventID:       event.EventID,
		Type:          constants.EventType(event.Type),
		OccurredAt:    event.OccurredAt,
		AccountID:     event.AccountID,
		ProcessedAt:   "",
		FailedAt:      "",
		LastErrorCode: "",
		NextRetryAt:   "",
	}
}
