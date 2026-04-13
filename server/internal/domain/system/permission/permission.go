package permission

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

type Verb string

func ToVerb(v string) (Verb, error) {
	verb := Verb(v)
	return verb, nil
}

type Permission struct {
	Id          string
	Verb        Verb
	ResourceId  string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func Factory(
	verb string,
	resourceId string,
	description string,
) (*Permission, error) {
	v, err := ToVerb(verb)
	if err != nil {
		return nil, err
	}
	return &Permission{
		Id:          uuid.MustUUID().String(),
		Verb:        v,
		ResourceId:  resourceId,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
