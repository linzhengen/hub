package token

import "time"

const AdminRoleName = "admin"

type Token struct {
	ClientId    string
	UserId      string
	AccessToken string
	//RefreshToken string
	Scope     string
	ExpiresAt time.Time
	//Revoked   bool
	Email             string
	EmailVerified     bool
	Roles             []string
	Username          string
	PreferredUsername string
}
