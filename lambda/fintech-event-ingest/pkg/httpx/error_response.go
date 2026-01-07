package httpx

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
	customerrors "github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/errors"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/logger"
)

type errorDetail struct {
	Code    constants.ErrorCode `json:"code"`
	Message string              `json:"message"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

const fallbackErrorJSON = `{"error":{"code":"internal_error","message":"Internal server error"}}`

// WithCorrelationID adds correlation ID to existing headers.
func WithCorrelationID(headers map[string]string, cid string) map[string]string {
	headers["X-Correlation-Id"] = cid
	return headers
}

// FromAppError constructs an APIGatewayProxyResponse from an AppError.
func FromAppError(
	ctx context.Context,
	req *events.APIGatewayProxyRequest,
	appErr *customerrors.AppError,
) events.APIGatewayProxyResponse {
	cid := CorrelationID(req)
	msg := constants.ErrorMessage(appErr.Code)

	attrs := []any{
		slog.String("correlation_id", cid),
		slog.String("error_code", string(appErr.Code)),
		slog.Bool("retryable", appErr.Retryable),
		slog.String("message", msg),
	}

	if appErr.Cause != nil {
		attrs = append(attrs, slog.String("cause", appErr.Cause.Error()))
	}

	logger.ErrorAttrs(ctx, "request failed", attrs...)

	body, err := json.Marshal(errorBody{
		Error: errorDetail{
			Code:    appErr.Code,
			Message: msg,
		},
	})
	if err != nil {
		slog.Error("failed to marshal error response", "error", err)

		body = []byte(fallbackErrorJSON)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:        appErr.HTTPStatus,
		Body:              string(body),
		Headers:           WithCorrelationID(constants.Headers(), cid),
		MultiValueHeaders: map[string][]string{},
		IsBase64Encoded:   false,
	}
}
