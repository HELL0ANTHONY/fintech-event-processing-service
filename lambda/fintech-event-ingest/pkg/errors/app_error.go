// Package customerrors defines a structured application error type for consistent error handling.
package customerrors

import (
	"net/http"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/lambda/fintech-event-ingest/pkg/constants"
)

// AppError represents a structured application error with additional context.
type AppError struct {
	Cause      error
	Code       constants.ErrorCode
	HTTPStatus int
	Retryable  bool
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}

	return string(e.Code)
}

// Validation creates a new AppError for validation errors.
func Validation(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorValidation,
		Cause:      err,
		Retryable:  false,
		HTTPStatus: http.StatusBadRequest,
	}
}

// DuplicateEvent creates a new AppError for duplicate event ID errors.
func DuplicateEvent() *AppError {
	return &AppError{
		Code:       constants.ErrorDuplicateEvent,
		Retryable:  false,
		HTTPStatus: http.StatusConflict,
		Cause:      nil,
	}
}

// Throttled creates a new AppError for resource throttling errors.
func Throttled(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorResourceThrottled,
		Cause:      err,
		Retryable:  true,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// Internal creates a new AppError for internal server errors.
func Internal(err error) *AppError {
	return &AppError{
		Code:       constants.ErrorInternal,
		Cause:      err,
		Retryable:  false,
		HTTPStatus: http.StatusInternalServerError,
	}
}
