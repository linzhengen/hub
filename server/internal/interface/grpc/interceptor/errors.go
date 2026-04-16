package interceptor

import (
	pkgError "github.com/linzhengen/hub/v1/server/pkg/error"
)

func TranslateError(err error) error {
	return pkgError.TranslateError(err)
}
