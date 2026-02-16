package handler

import (
	"context"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-processor/internal/processor"
)

// Handler processes DynamoDB Stream events.
type Handler struct {
	processor processor.StreamProcessor
}

// New creates a new Handler with the given StreamProcessor.
func New(p processor.StreamProcessor) *Handler {
	return &Handler{processor: p}
}

// Handle is the Lambda function handler that processes DynamoDB Stream events.
// It logs the number of records in the batch and delegates processing to the StreamProcessor.
func (h *Handler) Handle(ctx context.Context, event events.DynamoDBEvent) error {
	ctx = logger.With(ctx, "stream_event_count", len(event.Records))
	logger.Info(ctx, "processing stream batch")

	return h.processor.ProcessBatch(ctx, event.Records)
}
