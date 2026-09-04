// Package organization is the boundary a permission is held within.
//
// hub was single-tenant until this package existed: a grant meant the same
// thing everywhere, because `Enforce` matched a resource and an action and
// nothing else. An organization is the missing third term. A group belongs to
// one, so every route from a user to a permission passes through one, and a
// decision can be about a place rather than only about a verb.
//
// The package deliberately holds no notion of "the current organization". Which
// organization a request is about is a property of the request, resolved at the
// edge and carried in the context; storing it here would make it global state.
package organization

import (
	"regexp"
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

// PlatformOrgId is the organization that operates this installation.
//
// It is created by migration 000002 rather than by the seed, because
// `groups.org_id` is NOT NULL and the seed's own groups have to land somewhere.
// A schema whose first insert depends on `cli seed` having run is a schema that
// cannot be migrated on its own.
const PlatformOrgId = "00000000-0000-0000-0000-000000000001"

// Kind says what sort of tenant an organization is.
//
// Nothing branches on this to decide what is allowed - that is what Enforce and
// the graph are for. It exists so that a person reading a list can tell a
// company from an individual, and so that KindPlatform can carry the one rule
// below.
type Kind string

const (
	// KindPlatform is the operator of this installation. There is one.
	KindPlatform Kind = "PLATFORM"
	// KindBusiness is a company: many people, groups they define themselves.
	KindBusiness Kind = "BUSINESS"
	// KindPersonal is one person's own organization.
	//
	// An individual user is not a special case anywhere in hub; they are a
	// tenant of one. That is what lets a consumer travel the same authorization
	// path a company's staff do instead of needing a second one built for them.
	KindPersonal Kind = "PERSONAL"
)

// AppliesEverywhere reports whether a grant held through this kind of
// organization is answerable about any organization.
//
// Only the platform's is. This is the whole of the cross-organization rule, and
// it is stated once, here, rather than as a special case wherever a decision is
// made. It is also what makes the introduction of organizations safe: every
// group that existed before them was moved into the platform organization, so
// every grant that existed before them still means what it meant.
func (k Kind) AppliesEverywhere() bool {
	return k == KindPlatform
}

type Status string

const (
	Active   Status = "Active"
	InActive Status = "Inactive"
)

// Organization is one tenant.
type Organization struct {
	Id   string
	Name string
	// Slug is the handle an operator types and a URL carries. It is unique,
	// because it names the tenant in the places a UUID would be unreadable.
	Slug        string
	Kind        Kind
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// slugPattern is what a slug may be, and it is narrow for the same reason a
// service account's name is: it travels in URLs and in identifiers, so anything
// that would have to be escaped somewhere is refused here instead.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)

// ValidSlug reports whether slug may be used.
func ValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

// Platform reports whether this is the organization that operates the
// installation.
func (o *Organization) Platform() bool {
	return o.Kind.AppliesEverywhere()
}

// Factory builds an organization.
func Factory(name, slug string, kind Kind, description string) *Organization {
	now := time.Now()
	return &Organization{
		Id:          uuid.MustUUID().String(),
		Name:        name,
		Slug:        slug,
		Kind:        kind,
		Description: description,
		Status:      Active,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
