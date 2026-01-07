// Package processor implements the request processing logic for the Lambda function.
package processor

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/internal/normalization"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	customerrors "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/errors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/httpx"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/models"
)

// RequestProcessor defines the contract for processing incoming requests.
type RequestProcessor interface {
	Process(
		ctx context.Context,
		req *events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error)
}

// Processor implements the RequestProcessor interface.
type Processor struct {
	n normalization.Normalizable
}

// New creates a new RequestProcessor implementation.
func New(n normalization.Normalizable) RequestProcessor {
	return &Processor{
		n: n,
	}
}

// Process handles the business logic for the incoming request.
func (p *Processor) Process(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	transactions, err := models.NewJSONRequest(req.Body)
	if err != nil {
		appErr := customerrors.Validation(err)

		return httpx.FromAppError(ctx, req, appErr), nil
	}

	normalizedRequest := p.n.Normalize(transactions)

	logger.Debug(
		ctx,
		"normalized request",
		slog.Any("payload", normalizedRequest),
	)

	logger.Info(
		ctx,
		"request processed successfully",
		slog.Int("transaction_count", len(normalizedRequest.Event)),
	)

	return events.APIGatewayProxyResponse{
		StatusCode:        http.StatusOK,
		Body:              "Request processed successfully",
		IsBase64Encoded:   false,
		Headers:           constants.Headers(),
		MultiValueHeaders: map[string][]string{},
	}, nil
}
