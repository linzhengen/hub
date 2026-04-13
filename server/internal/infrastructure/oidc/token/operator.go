package token

import (
	"context"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	oidctoken "github.com/linzhengen/hub/server/internal/domain/oidc/token"
	jwtutil "github.com/linzhengen/hub/server/pkg/jwt"
)

type operatorImpl struct {
	keyCloak *KeyCloak
}

type KeyCloak struct {
	ClientId    string
	Realm       string
	Client      *gocloak.GoCloak
	DisableAuth bool
}

func New(keyCloak *KeyCloak) oidctoken.Operator {
	return &operatorImpl{
		keyCloak: keyCloak,
	}
}

func (o operatorImpl) ValidateToken(ctx context.Context, tokenString string) (*oidctoken.Token, error) {
	_, c, err := o.keyCloak.Client.DecodeAccessToken(ctx, tokenString, o.keyCloak.Realm)
	if err != nil {
		return nil, err
	}
	uid, err := c.GetSubject()
	if err != nil {
		return nil, err
	}
	et, err := c.GetExpirationTime()
	if err != nil {
		return nil, err
	}
	return &oidctoken.Token{
		ClientId:          o.keyCloak.ClientId,
		UserId:            uid,
		AccessToken:       tokenString,
		Scope:             o.getScope(c),
		ExpiresAt:         et.Time,
		Email:             o.getEmail(c),
		EmailVerified:     false,
		Roles:             o.getRoles(c),
		Username:          o.getName(c),
		PreferredUsername: o.getPreferredUsername(c),
	}, nil
}

func (o operatorImpl) ExtractToken(ctx context.Context) (*oidctoken.Token, error) {
	if o.keyCloak.DisableAuth {
		return nil, nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, oidctoken.ErrNoSession
	}
	t, err := o.getToken(md)
	if err != nil {
		return nil, err
	}
	return o.ValidateToken(ctx, t)
}

func (o operatorImpl) getToken(md metadata.MD) (string, error) {
	// check the "token" metadata
	{
		tokens, ok := md["token"]
		if ok && len(tokens) > 0 {
			return tokens[0], nil
		}
	}

	// looks for the HTTP header `Authorization: Bearer ...`
	for _, t := range md["authorization"] {
		tt := strings.TrimPrefix(t, "Bearer ")
		if strings.HasPrefix(t, "Bearer ") && jwtutil.IsValid(tt) {
			return tt, nil
		}
	}

	return "", oidctoken.ErrNoSession
}

func (o operatorImpl) getRoles(c *jwt.MapClaims) []string {
	if c == nil {
		return nil
	}

	realmAccess := getClaimValue[map[string]interface{}](c, "realm_access")
	if realmAccess == nil {
		return nil
	}

	roles, ok := realmAccess["roles"].([]interface{})
	if !ok {
		return nil
	}

	var r []string
	for _, role := range roles {
		roleStr, ok := role.(string)
		if ok {
			r = append(r, roleStr)
		}
	}
	return r
}

func getClaimValue[T any](c *jwt.MapClaims, key string) T {
	var zero T
	if c == nil {
		return zero
	}

	value, ok := (*c)[key].(T)
	if !ok {
		return zero
	}
	return value
}

func (o operatorImpl) getScope(c *jwt.MapClaims) string {
	return getClaimValue[string](c, "scope")
}

func (o operatorImpl) getEmail(c *jwt.MapClaims) string {
	return getClaimValue[string](c, "email")
}

func (o operatorImpl) getPreferredUsername(c *jwt.MapClaims) string {
	return getClaimValue[string](c, "preferred_username")
}

func (o operatorImpl) getName(c *jwt.MapClaims) string {
	return getClaimValue[string](c, "name")
}
