package constants

// ErrorCode defines standardized error codes for the service.
type ErrorCode string

const (
	// Validation errors.
	ErrValidation       ErrorCode = "validation_error"
	ErrInvalidSchema    ErrorCode = "invalid_schema"
	ErrInvalidAmount    ErrorCode = "invalid_amount"
	ErrInvalidCurrency  ErrorCode = "invalid_currency"
	ErrInvalidEventType ErrorCode = "invalid_event_type"
	ErrInvalidTimestamp ErrorCode = "invalid_timestamp"
	ErrMissingField     ErrorCode = "missing_required_field"
	ErrPayloadTooLarge  ErrorCode = "payload_too_large"
	ErrBatchTooLarge    ErrorCode = "batch_too_large"

	// Business rule errors.
	ErrBusinessRule          ErrorCode = "business_rule_violation"
	ErrInvalidMovement       ErrorCode = "invalid_movement"
	ErrInvalidCredit         ErrorCode = "invalid_credit"
	ErrMissingReference      ErrorCode = "missing_reference"
	ErrInvalidMatchStatus    ErrorCode = "invalid_match_status"
	ErrMissingMismatchReason ErrorCode = "missing_mismatch_reason"

	// Idempotency errors.
	ErrDuplicateEvent ErrorCode = "duplicate_event_id"
	ErrConflict       ErrorCode = "conflict"

	// Transient errors (retryable).
	ErrTimeout           ErrorCode = "timeout"
	ErrResourceThrottled ErrorCode = "resource_throttled"
	ErrDependencyFailed  ErrorCode = "dependency_failed"
	ErrRetryExhausted    ErrorCode = "retry_exhausted"

	// System errors.
	ErrInternal     ErrorCode = "internal_error"
	ErrNotFound     ErrorCode = "not_found"
	ErrExportFailed ErrorCode = "export_failed"
)

var errorMessages = map[ErrorCode]string{
	ErrValidation:            "The request payload is invalid or violates validation rules",
	ErrInvalidSchema:         "The request does not conform to the expected schema",
	ErrInvalidAmount:         "The amount value is invalid",
	ErrInvalidCurrency:       "The currency is not supported",
	ErrInvalidEventType:      "The event type is not supported",
	ErrInvalidTimestamp:      "The timestamp format is invalid or out of range",
	ErrMissingField:          "A required field is missing",
	ErrPayloadTooLarge:       "The request payload exceeds the maximum allowed size",
	ErrBatchTooLarge:         "The number of events exceeds the maximum allowed per request",
	ErrBusinessRule:          "The request violates business rules",
	ErrInvalidMovement:       "The movement event contains invalid data",
	ErrInvalidCredit:         "The credit event contains invalid data",
	ErrMissingReference:      "Credit settlement requires a reference",
	ErrInvalidMatchStatus:    "The reconciliation match status is invalid",
	ErrMissingMismatchReason: "Mismatched reconciliation requires a reason",
	ErrDuplicateEvent:        "Event already processed or currently being processed",
	ErrConflict:              "The request conflicts with the current resource state",
	ErrTimeout:               "Processing exceeded the allowed time limit",
	ErrResourceThrottled:     "Service temporarily unavailable due to high load",
	ErrDependencyFailed:      "A dependent service is temporarily unavailable",
	ErrRetryExhausted:        "Maximum retry attempts exceeded",
	ErrInternal:              "An unexpected internal error occurred",
	ErrExportFailed:          "Failed to complete export operation",
	ErrNotFound:              "The requested resource was not found",
}

// Message returns the human-readable message for an error code.
func (e ErrorCode) Message() string {
	if msg, ok := errorMessages[e]; ok {
		return msg
	}

	return "Unknown error"
}

// IsRetryable returns true if the error is transient and can be retried.
func (e ErrorCode) IsRetryable() bool {
	switch e {
	case ErrTimeout, ErrResourceThrottled, ErrDependencyFailed:
		return true
	default:
		return false
	}
}
