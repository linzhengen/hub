package interceptor

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TranslateError(err error) error {
	switch err {
	case nil:
		return nil
	default:
		// TODO: handle error translation
		return status.Error(codes.Unknown, err.Error())
	}
}
