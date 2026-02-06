package httpx

import (
	"encoding/json"
	"maps"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/customerrors"
	"github.com/aws/aws-lambda-go/events"
)

// ErrorDetail represents the error structure in API responses.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorBody is the top-level error response structure.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

const fallbackErrorJSON = `{"error":{"code":"internal_error","message":"Internal server error"}}`

// WithCorrelationID adds correlation ID to the headers.
func WithCorrelationID(headers map[string]string, cid string) map[string]string {
	result := make(map[string]string)
	maps.Copy(result, headers)
	result["X-Correlation-Id"] = cid

	return result
}

// ErrorResponse constructs an API Gateway response from an AppError.
func ErrorResponse(
	appErr *customerrors.AppError,
	correlationID string,
) events.APIGatewayProxyResponse {
	body, err := json.Marshal(ErrorBody{
		Error: ErrorDetail{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		},
	})
	if err != nil {
		body = []byte(fallbackErrorJSON)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      appErr.HTTPStatus,
		Body:            string(body),
		Headers:         WithCorrelationID(constants.StandardHeaders(), correlationID),
		IsBase64Encoded: false,
	}
}

// SuccessResponse constructs a successful API Gateway response.
func SuccessResponse(
	statusCode int,
	body any,
	correlationID string,
) events.APIGatewayProxyResponse {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return ErrorResponse(
			customerrors.Internal(err),
			correlationID,
		)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Body:            string(jsonBody),
		Headers:         WithCorrelationID(constants.StandardHeaders(), correlationID),
		IsBase64Encoded: false,
	}
}
