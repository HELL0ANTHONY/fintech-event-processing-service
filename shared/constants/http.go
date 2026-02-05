// Package constants contains constant definitions used across the fintech event processing service.
package constants

import "net/http"

// StandardHeaders returns common HTTP headers for API responses.
func StandardHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "POST, GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, X-Correlation-Id, X-Idempotency-Key",
	}
}

// HTTPStatusFromErrorCode maps error codes to HTTP status codes.
func HTTPStatusFromErrorCode(code ErrorCode) int {
	switch code {
	// 400 Bad Request - Validation errors.
	case ErrValidation, ErrInvalidSchema, ErrInvalidAmount, ErrInvalidCurrency,
		ErrInvalidEventType, ErrInvalidTimestamp, ErrMissingField,
		ErrPayloadTooLarge, ErrBatchTooLarge, ErrBusinessRule,
		ErrInvalidMovement, ErrInvalidCredit, ErrMissingReference,
		ErrInvalidMatchStatus, ErrMissingMismatchReason:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrDuplicateEvent, ErrConflict:
		return http.StatusConflict
	case ErrResourceThrottled:
		return http.StatusTooManyRequests
	case ErrTimeout:
		return http.StatusGatewayTimeout
	case ErrDependencyFailed:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
