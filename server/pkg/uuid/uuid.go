package uuid

import "github.com/google/uuid"

type UUID = uuid.UUID

func NewUUID() (UUID, error) {
	return uuid.NewV7()
}

func MustUUID() UUID {
	return uuid.Must(NewUUID())
}

func MustString() string {
	return MustUUID().String()
}
