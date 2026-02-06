package handler

import (
	"context"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/logger"
	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/httpx"
)

// Handler processes incoming API Gateway requests.
type Handler struct {
	processor processor.RequestProcessor
}

// New creates a new Handler with the given processor.
func New(p processor.RequestProcessor) *Handler {
	return &Handler{processor: p}
}

// Handle is the AWS Lambda entry point.
func (h *Handler) Handle(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	correlationID := httpx.CorrelationID(req)
	ctx = logger.WithRequestContext(ctx, req.RequestContext.RequestID, correlationID, req.Path)

	return h.processor.Process(ctx, req, correlationID)
}
