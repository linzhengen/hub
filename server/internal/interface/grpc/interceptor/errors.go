package interceptor

import (
	pkgError "github.com/linzhengen/hub/server/pkg/error"
)

func TranslateError(err error) error {
	return pkgError.TranslateError(err)
}
