// Package handler implements the AWS Lambda handler for processing fintech events.
package handler

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/processor"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

type Handler struct {
	p processor.RequestProcessor
}

// New creates a new Handler with the given RequestProcessor.
func New(p processor.RequestProcessor) Handler {
	return Handler{
		p: p,
	}
}

// Handle is the AWS Lambda handler function.
func (h Handler) Handle(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	ctx = logger.With(
		ctx,
		slog.String("request_id", req.RequestContext.RequestID),
		slog.String("path", req.Path),
	)

	return h.p.Process(ctx, req)
}
