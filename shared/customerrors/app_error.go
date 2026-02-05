// Package customerrors provides structured application error types.
package customerrors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/aws/smithy-go"

	"github.com/HELL0ANTHONY/fintech-event-processing-service/shared/constants"
)

// AppError represents a structured application error with context.
type AppError struct {
	Cause      error
	Code       constants.ErrorCode
	Message    string
	EventID    string
	HTTPStatus int
	Retryable  bool
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Validation creates a validation error.
func Validation(code constants.ErrorCode, cause error) *AppError {
	return newAppError(code, code.Message(), cause, false)
}

// ValidationWithMessage creates a validation error with a custom message.
func ValidationWithMessage(code constants.ErrorCode, msg string) *AppError {
	return newAppError(code, msg, nil, false)
}

// BusinessRule creates a business rule violation error.
func BusinessRule(code constants.ErrorCode, msg string) *AppError {
	return newAppError(code, msg, nil, false)
}

// Duplicate creates a duplicate event error.
func Duplicate(eventID string) *AppError {
	return newAppError(
		constants.ErrDuplicateEvent,
		constants.ErrDuplicateEvent.Message(),
		nil,
		false,
		withEventID(eventID),
	)
}

// NotFound creates a not found error for a given resource.
func NotFound(resource, id string) *AppError {
	msg := resource + " not found"
	if id != "" {
		msg = fmt.Sprintf("%s %s not found", resource, id)
	}

	return newAppError(
		constants.ErrNotFound,
		msg,
		nil,
		false,
	)
}

// Throttled creates a throttling error.
func Throttled(cause error) *AppError {
	return newAppError(
		constants.ErrResourceThrottled,
		constants.ErrResourceThrottled.Message(),
		cause,
		true,
	)
}

// Timeout creates a timeout error.
func Timeout(cause error) *AppError {
	return newAppError(constants.ErrTimeout, constants.ErrTimeout.Message(), cause, true)
}

// DependencyFailed creates a dependency failure error.
func DependencyFailed(cause error) *AppError {
	return newAppError(
		constants.ErrDependencyFailed,
		constants.ErrDependencyFailed.Message(),
		cause,
		true,
	)
}

// RetryExhausted creates a retry exhausted error.
func RetryExhausted(eventID string, attempts int) *AppError {
	return newAppError(
		constants.ErrRetryExhausted,
		fmt.Sprintf("event %s failed after %d attempts", eventID, attempts),
		nil,
		false,
		withEventID(eventID),
	)
}

// Internal creates an internal server error.
func Internal(cause error) *AppError {
	return newAppError(constants.ErrInternal, constants.ErrInternal.Message(), cause, false)
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}

	return false
}

// GetErrorCode extracts the error code from an error.
func GetErrorCode(err error) constants.ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	return constants.ErrInternal
}

// WrapTransient wraps an error as transient if it matches known transient patterns.
// Uses type-safe error checking where possible, falling back to string matching
// for AWS-specific errors that don't expose typed errors.
func WrapTransient(err error) *AppError {
	if err == nil {
		return nil
	}

	// 1. Context errors (type-safe).
	if errors.Is(err, context.DeadlineExceeded) {
		return Timeout(err)
	}

	if errors.Is(err, context.Canceled) {
		return Timeout(err)
	}

	// 2. Network errors (type-safe).
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return Timeout(err)
		}

		return DependencyFailed(err)
	}

	// 3. AWS SDK errors (type-safe where possible).
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return wrapAWSError(apiErr, err)
	}

	// 4. Fallback to string matching for edge cases.
	return wrapByErrorString(err)
}

// wrapAWSError handles AWS SDK v2 errors using the smithy error interface.
func wrapAWSError(apiErr smithy.APIError, originalErr error) *AppError {
	code := apiErr.ErrorCode()

	// Throttling errors.
	throttlingCodes := []string{
		"ProvisionedThroughputExceededException",
		"ThrottlingException",
		"RequestLimitExceeded",
		"Throttling",
		"TooManyRequestsException",
	}

	if slices.Contains(throttlingCodes, code) {
		return Throttled(originalErr)
	}

	// Transient/retryable errors.
	transientCodes := []string{
		"ServiceUnavailable",
		"InternalServerError",
		"RequestTimeout",
		"IDPCommunicationError",
	}

	if slices.Contains(transientCodes, code) {
		return DependencyFailed(originalErr)
	}

	// Conditional check failure (duplicate) - not transient.
	if code == "ConditionalCheckFailedException" {
		return Duplicate("")
	}

	return Internal(originalErr)
}

// wrapByErrorString is a fallback for errors that don't implement typed interfaces.
func wrapByErrorString(err error) *AppError {
	errStr := err.Error()

	switch {
	case containsAny(errStr, "connection refused", "no such host", "network unreachable"):
		return DependencyFailed(err)
	case containsAny(errStr, "timeout", "Timeout"):
		return Timeout(err)
	default:
		return Internal(err)
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}

	return false
}

func newAppError(
	code constants.ErrorCode,
	msg string,
	cause error,
	retryable bool,
	opts ...func(*AppError),
) *AppError {
	err := &AppError{
		Code:       code,
		Message:    msg,
		Cause:      cause,
		HTTPStatus: constants.HTTPStatusFromErrorCode(code),
		Retryable:  retryable,
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

func withEventID(eventID string) func(*AppError) {
	return func(e *AppError) {
		e.EventID = eventID
	}
}
