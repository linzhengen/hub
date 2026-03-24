package error

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError represents a validation error with field-specific details.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	for i, err := range e {
		_, _ = fmt.Fprintf(&sb, "  %s: %s", err.Field, err.Message)
		if i < len(e)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// NewValidationError creates a new validation error for a single field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrors creates a collection of validation errors.
func NewValidationErrors(errors ...ValidationError) ValidationErrors {
	return ValidationErrors(errors)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var validationError *ValidationError
	var validationErrors ValidationErrors
	switch {
	case errors.As(err, &validationError), errors.As(err, &validationErrors):
		return true
	default:
		return false
	}
}
