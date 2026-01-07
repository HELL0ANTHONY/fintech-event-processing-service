package models

import "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"

// DynamoDBEvent represents the structure of an event stored in DynamoDB.
type DynamoDBEvent struct {
	Metadata      map[string]any        `json:"metadata"                dynamodbav:"metadata"`
	Amount        Amount                `json:"amount"                  dynamodbav:"amount"`
	Status        constants.EventStatus `json:"status"                  dynamodbav:"status"`
	ReceivedAt    string                `json:"receivedAt"              dynamodbav:"receivedAt"`
	ProcessedAt   string                `json:"processedAt,omitempty"   dynamodbav:"processedAt,omitempty"`
	FailedAt      string                `json:"failedAt,omitempty"      dynamodbav:"failedAt,omitempty"`
	LastErrorCode string                `json:"lastErrorCode,omitempty" dynamodbav:"lastErrorCode,omitempty"`
	NextRetryAt   string                `json:"nextRetryAt,omitempty"   dynamodbav:"nextRetryAt,omitempty"`
	BatchID       string                `json:"batchId"                 dynamodbav:"batchId"`
	EventID       string                `json:"eventId"                 dynamodbav:"eventId"`
	Type          constants.EventType   `json:"type"                    dynamodbav:"type"`
	OccurredAt    string                `json:"occurredAt"              dynamodbav:"occurredAt"`
	AccountID     string                `json:"accountId"               dynamodbav:"accountId"`
	Source        constants.Source      `json:"source"                  dynamodbav:"source"`
	Attempts      int                   `json:"attempts"                dynamodbav:"attempts"`
}

// NewDynamoDBEventFromRequest creates a new DynamoDBEvent from a Request.
func NewDynamoDBEventFromRequest(
	event *Event,
	source constants.Source,
	receivedAt, batchID string,
) *DynamoDBEvent {
	return &DynamoDBEvent{
		Source:     source,
		BatchID:    batchID,
		ReceivedAt: receivedAt,
		Status:     constants.StatusReceived,
		Attempts:   0,
		Metadata:   event.Metadata,
		Amount:     event.Amount,
		EventID:    event.EventID,
		Type:       constants.EventType(event.Type),
		OccurredAt: event.OccurredAt,
		AccountID:  event.AccountID,
	}
}
