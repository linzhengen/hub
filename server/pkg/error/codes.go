package error

// Common error codes
const (
	// CodeInvalidArgument indicates invalid argument provided
	CodeInvalidArgument = "invalid_argument"
	// CodeNotFound indicates resource not found
	CodeNotFound = "not_found"
	// CodeAlreadyExists indicates resource already exists
	CodeAlreadyExists = "already_exists"
	// CodePermissionDenied indicates permission denied
	CodePermissionDenied = "permission_denied"
	// CodeUnauthenticated indicates authentication required
	CodeUnauthenticated = "unauthenticated"
	// CodeResourceExhausted indicates resource exhausted (e.g., rate limit)
	CodeResourceExhausted = "resource_exhausted"
	// CodeFailedPrecondition indicates precondition failed
	CodeFailedPrecondition = "failed_precondition"
	// CodeAborted indicates operation aborted
	CodeAborted = "aborted"
	// CodeInternal indicates internal server error
	CodeInternal = "internal"
	// CodeUnavailable indicates service unavailable
	CodeUnavailable = "unavailable"
	// CodeValidationFailed indicates validation failed
	CodeValidationFailed = "validation_failed"
)
