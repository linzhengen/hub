package user

type User struct {
	Name           string
	Email          string
	EmailVerified  bool
	HashedPassword string
}
