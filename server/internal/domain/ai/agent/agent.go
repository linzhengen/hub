// Package agent is the register of the programs that call hub's API.
//
// An agent acts **as itself**. It authenticates as its own Keycloak client and
// is authorized against its own grants; nothing here lets it claim to be acting
// for a person, because the `delegations` row that would make such a claim
// checkable does not exist yet. Until it does, an agent is a service account
// with a tenant and a lineage - and an audit record of what it did names the
// agent, not whoever asked it.
//
// Authentication stays Keycloak's: an agent is a Keycloak client and it gets a
// token through the client credentials grant, exactly as a service account
// does. What this package adds is that hub knows the client is there - which
// tenant it belongs to, what it is for, who set it up, which agent owns it, and
// which hub user it acts as.
//
// The user is the load-bearing part. An agent holding a `users` row joins
// groups, holds roles, takes time-bounded grants and appears in the audit log
// through the machinery a person does, so there is one authorization path
// rather than two that have to be kept in step.
package agent

import (
	"regexp"
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

// AuthMethod is how an agent proves it is its Keycloak client.
//
// The method is a column so that a shared secret and a signed assertion can
// coexist while agents are moved across, rather than needing a flag day.
// ADR: docs/decisions/2026-09-04-keep-the-agent-credential-a-client-secret-and-record-its-method.md
type AuthMethod string

const (
	// AuthMethodClientSecret is a shared secret, issued once and rotated on
	// demand. It is the only method hub issues today.
	AuthMethodClientSecret AuthMethod = "CLIENT_SECRET"
	// AuthMethodPrivateKeyJWT is a signed assertion against a key the client
	// holds, so no shared secret exists to leak. Keycloak supports it and hub
	// does not issue it yet; the constant exists so that the column's meaning
	// is stated in one place when it does.
	AuthMethodPrivateKeyJWT AuthMethod = "PRIVATE_KEY_JWT"
)

// Agent is one program's identity.
type Agent struct {
	Id string
	// OrgId is the tenant the agent belongs to. Every route from the agent to a
	// permission passes through an organization already; this says which one
	// the agent itself is a part of.
	OrgId string
	// UserId is the hub user the agent acts as - the Keycloak client's own
	// service account user, stored in `users` like any other.
	UserId      string
	Name        string
	Description string
	// ClientId is what an operator configures; KeycloakId is Keycloak's
	// internal handle, which every admin call needs.
	ClientId   string
	KeycloakId string
	AuthMethod AuthMethod
	// ParentAgentId is the agent that owns this one, empty for a root agent.
	//
	// Nothing is enforced on it yet. A sub-agent's authority will be bounded by
	// the delegation chain rather than by this column, which records lineage so
	// that the chain has something to walk.
	ParentAgentId string
	// CreatedByUserId is the human this agent answers to. It is not provenance
	// trivia: an agent with no identifiable controller is the one nobody turns
	// off, so the reference is kept alive by the schema.
	CreatedByUserId string
	// SecretRotatedAt is when the current credential was issued. hub never
	// stores the credential, so its age is the only thing that can be reported
	// about it.
	SecretRotatedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Root reports whether the agent is at the top of its lineage.
func (a *Agent) Root() bool {
	return a.ParentAgentId == ""
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
// It is derived rather than chosen so that two agents cannot claim the same
// client, and prefixed so that a client hub created is recognisable among
// clients it did not - deleting one it did not create is a mistake worth making
// impossible to reach by accident. The prefix also separates agents from
// service accounts, which take `hub-sa-`.
func ClientIdFor(name string) string {
	return "hub-agent-" + name
}

// UsernameFor is the username the agent's hub user carries.
//
// Keycloak names a client's service account user `service-account-<client id>`
// and hub follows it, so the row in `users` reads the same as the user Keycloak
// shows and neither has to be translated into the other.
func UsernameFor(name string) string {
	return "service-account-" + ClientIdFor(name)
}

// EmailFor is the address the agent's hub user carries.
//
// `users.email` is unique and not null, and an agent has no mailbox. The
// address is therefore synthesised from the name - unique because the name is -
// and points at a domain reserved by RFC 2606 so that nothing ever tries to
// deliver to it.
func EmailFor(name string) string {
	return ClientIdFor(name) + "@service-account.invalid"
}

// Factory builds an agent from what registering its Keycloak client returned.
func Factory(
	orgId, name, description, clientId, keycloakId, userId, parentAgentId, createdByUserId string,
) *Agent {
	now := time.Now()
	return &Agent{
		Id:              uuid.MustUUID().String(),
		OrgId:           orgId,
		UserId:          userId,
		Name:            name,
		Description:     description,
		ClientId:        clientId,
		KeycloakId:      keycloakId,
		AuthMethod:      AuthMethodClientSecret,
		ParentAgentId:   parentAgentId,
		CreatedByUserId: createdByUserId,
		SecretRotatedAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
