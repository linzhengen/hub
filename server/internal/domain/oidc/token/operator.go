package token

import "context"

type Operator interface {
	ValidateToken(ctx context.Context, accessToken string) (*Token, error)
	ExtractToken(ctx context.Context) (*Token, error)
}
