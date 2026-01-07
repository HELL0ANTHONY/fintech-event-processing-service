// Package httpx provides HTTP-related utilities for AWS Lambda functions.
package httpx

import "github.com/aws/aws-lambda-go/events"

// CorrelationID extracts the correlation ID from the request headers.
func CorrelationID(req *events.APIGatewayProxyRequest) string {
	if id := req.Headers["X-Correlation-Id"]; id != "" {
		return id
	}

	return req.RequestContext.RequestID
}
