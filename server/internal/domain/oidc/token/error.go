package token

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrNoSession = status.Errorf(codes.Unauthenticated, "no session information")
var ErrInvalidToken = status.Errorf(codes.Unauthenticated, "invalid token")
