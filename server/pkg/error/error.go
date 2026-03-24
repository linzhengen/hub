package error

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error represents a structured application error.
type Error struct {
	// Code is a machine-readable error code.
	Code string
	// Message is a human-readable error message.
	Message string
	// Cause is the underlying error that triggered this error.
	Cause error
	// Metadata contains additional key-value pairs for context.
	Metadata map[string]any
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error for error wrapping compatibility.
func (e *Error) Unwrap() error {
	return e.Cause
}

// GRPCCode returns the gRPC status code that best represents this error.
func (e *Error) GRPCCode() codes.Code {
	// Map common error codes to gRPC codes
	switch e.Code {
	case CodeInvalidArgument, CodeValidationFailed:
		return codes.InvalidArgument
	case CodeNotFound:
		return codes.NotFound
	case CodeAlreadyExists:
		return codes.AlreadyExists
	case CodePermissionDenied:
		return codes.PermissionDenied
	case CodeUnauthenticated:
		return codes.Unauthenticated
	case CodeResourceExhausted:
		return codes.ResourceExhausted
	case CodeFailedPrecondition:
		return codes.FailedPrecondition
	case CodeAborted:
		return codes.Aborted
	case CodeInternal:
		return codes.Internal
	case CodeUnavailable:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

// New creates a new Error with the given code and message.
func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new Error with the given code and formatted message.
func Newf(code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with a code and message.
func Wrap(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with a code and formatted message.
func Wrapf(err error, code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// WithMetadata adds metadata to the error.
func (e *Error) WithMetadata(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// TranslateError translates an error to a gRPC status error.
// It handles both structured Error types and standard Go errors.
func TranslateError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already a gRPC status error, return it as-is
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Handle structured Error type
	var appErr *Error
	if errors.As(err, &appErr) {
		return status.Error(appErr.GRPCCode(), appErr.Message)
	}

	// Handle common error types
	// Database errors
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.NotFound, "resource not found")
	}

	// Domain errors (check by error string comparison)
	errStr := err.Error()
	switch errStr {
	case "unauthorized":
		return status.Error(codes.Unauthenticated, "unauthorized")
	case "invalid request":
		return status.Error(codes.InvalidArgument, "invalid request")
	}

	// Validation errors
	if IsValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Return unknown for unmapped errors
	return status.Error(codes.Unknown, err.Error())
}

// Is checks if the target error is an Error with the given code.
func Is(err error, code string) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
