package constants

// ErrorCode defines the types of error codes used in the fintech event ingest service.
type ErrorCode string

// Predefined error codes for various error scenarios.
const (
	ErrorValidation     ErrorCode = "validation_error"
	ErrorDuplicateEvent ErrorCode = "duplicate_event_id"
	ErrorConflict       ErrorCode = "conflict"

	ErrorTimeout           ErrorCode = "timeout"
	ErrorResourceThrottled ErrorCode = "resource_throttled"
	ErrorExportFailed      ErrorCode = "export_failed"

	ErrorInternal ErrorCode = "internal_error"
)
