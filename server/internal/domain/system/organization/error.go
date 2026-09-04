package organization

import "errors"

var (
	// ErrInvalidSlug is a slug that would need escaping somewhere it travels.
	ErrInvalidSlug = errors.New("invalid organization slug")
	// ErrPlatformImmutable is an attempt to delete or re-kind the organization
	// that operates the installation. Removing it would orphan every group that
	// predates organizations and silently revoke every administrator.
	ErrPlatformImmutable = errors.New("the platform organization cannot be deleted or changed in kind")
)
