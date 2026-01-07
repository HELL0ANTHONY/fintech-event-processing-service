package constants

var errorMessages = map[ErrorCode]string{
	ErrorConflict:          "The request conflicts with the current resource state",
	ErrorDuplicateEvent:    "Event already processed or currently being processed",
	ErrorExportFailed:      "Failed to complete export operation",
	ErrorInternal:          "Unexpected internal error occurred",
	ErrorResourceThrottled: "Service temporarily unavailable due to high load",
	ErrorTimeout:           "Processing exceeded the allowed time limit",
	ErrorValidation:        "The request payload is invalid or violates business rules",
}

// ErrorMessage returns the error message corresponding to the given ErrorCode.
func ErrorMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}

	return "Unknown error"
}
