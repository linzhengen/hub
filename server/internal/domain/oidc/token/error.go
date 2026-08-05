package token

import "errors"

var ErrNoSession = errors.New("no session information")
var ErrInvalidToken = errors.New("invalid token")
