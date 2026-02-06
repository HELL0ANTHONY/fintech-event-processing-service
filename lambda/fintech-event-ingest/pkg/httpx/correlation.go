package httpx

import "github.com/aws/aws-lambda-go/events"

// CorrelationID extracts the correlation ID from request headers.
// Falls back to the request ID if no correlation ID header is present.
func CorrelationID(req *events.APIGatewayProxyRequest) string {
	if id := req.Headers["X-Correlation-Id"]; id != "" {
		return id
	}

	if id := req.Headers["x-correlation-id"]; id != "" {
		return id
	}

	return req.RequestContext.RequestID
}
