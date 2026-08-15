// Package serviceaccount is the register of machines that call the API.
//
// Authentication stays Keycloak's: a service account is a Keycloak client, and
// a machine gets a token through the client credentials grant exactly as it did
// before this package existed. What is added is that hub knows the client is
// there - what it is for, who set it up, and which hub user it acts as.
package serviceaccount

import (
	"regexp"
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

// ServiceAccount is one machine's identity.
type ServiceAccount struct {
	Id string
	// UserId is the hub user the machine acts as - the Keycloak client's own
	// service account user, stored in `users` like any other.
	//
	// This is what keeps the rest of hub unchanged: a machine joins groups,
	// holds roles and is recorded in the audit log through the same machinery a
	// person is, rather than through a second authorization path that would
	// have to be kept in step with the first.
	UserId string
	Name   string
	// Description says what the machine is for. A machine nobody can describe
	// is a machine nobody dares turn off.
	Description string
	// ClientId is what an operator puts in HUB_OIDC_CLIENT_ID; KeycloakId is
	// Keycloak's internal handle, which every admin call needs.
	ClientId        string
	KeycloakId      string
	CreatedByUserId string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// namePattern is what a name may be, and it is deliberately narrow: the name
// becomes a Keycloak client id and a generated username, so anything that would
// need escaping somewhere is refused here instead.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)

// ValidName reports whether name may be used.
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

// ClientIdFor is the Keycloak client id a name maps to.
//
// It is derived rather than chosen so that two service accounts cannot claim
// the same client, and prefixed so that a client hub created is recognisable
// among clients it did not - deleting one it did not create is a mistake worth
// making impossible to reach by accident.
func ClientIdFor(name string) string {
	return "hub-sa-" + name
}

// UsernameFor is the username the machine's hub user carries.
//
// Keycloak names a client's service account user `service-account-<client id>`
// and hub follows it, so the row in `users` reads the same as the user Keycloak
// shows and neither has to be translated into the other.
func UsernameFor(name string) string {
	return "service-account-" + ClientIdFor(name)
}

// EmailFor is the address the machine's hub user carries.
//
// `users.email` is unique and not null, and a machine has no mailbox. The
// address is therefore synthesised from the name - unique because the name is -
// and points at a domain reserved by RFC 2606 so that nothing ever tries to
// deliver to it.
func EmailFor(name string) string {
	return ClientIdFor(name) + "@service-account.invalid"
}

// Factory builds a service account from what registering its Keycloak client
// returned.
func Factory(name, description, clientId, keycloakId, userId, createdByUserId string) *ServiceAccount {
	now := time.Now()
	return &ServiceAccount{
		Id:              uuid.MustUUID().String(),
		UserId:          userId,
		Name:            name,
		Description:     description,
		ClientId:        clientId,
		KeycloakId:      keycloakId,
		CreatedByUserId: createdByUserId,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
