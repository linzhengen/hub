package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The name becomes a Keycloak client id and a username, so it is validated
// narrowly here rather than escaped in three places later.
func TestValidName(t *testing.T) {
	valid := []string{"support-triage", "nightly-report-2", "a1b", "ops"}
	for _, name := range valid {
		assert.True(t, ValidName(name), "%q should be accepted", name)
	}

	invalid := map[string]string{
		"ab":                    "two characters is too short to describe anything",
		"-triage":               "leading hyphen",
		"triage-":               "trailing hyphen",
		"Triage":                "uppercase would not survive a client id",
		"support triage":        "a space",
		"support/triage":        "a path separator",
		"support@triage":        "an address separator",
		strings.Repeat("a", 65): "longer than the column",
		"":                      "empty",
	}
	for name, why := range invalid {
		assert.False(t, ValidName(name), "%q should be refused: %s", name, why)
	}
}

// The derived identifiers are what tie hub's row, Keycloak's client and the
// agent's user together. They are derived rather than chosen so that two agents
// cannot claim the same client.
func TestDerivedIdentifiers(t *testing.T) {
	assert.Equal(t, "hub-agent-support-triage", ClientIdFor("support-triage"))
	assert.Equal(t, "service-account-hub-agent-support-triage", UsernameFor("support-triage"))

	// `users.email` is unique and not null and an agent has no mailbox, so the
	// address is synthesised. `.invalid` is reserved by RFC 2606, so nothing
	// ever tries to deliver to it.
	assert.Equal(t, "hub-agent-support-triage@service-account.invalid", EmailFor("support-triage"))
	assert.True(t, strings.HasSuffix(EmailFor("anything"), ".invalid"))

	// Two names never collide, which is what keeps the unique columns honest.
	assert.NotEqual(t, ClientIdFor("triage"), ClientIdFor("triage-2"))
	assert.NotEqual(t, EmailFor("triage"), EmailFor("triage-2"))
}

// An agent and a service account must never derive the same Keycloak client
// from the same name. They are different registers with different lifecycles,
// and one deleting the other's client would be unrecoverable.
func TestAgentClientsDoNotCollideWithServiceAccountClients(t *testing.T) {
	assert.Equal(t, "hub-agent-triage", ClientIdFor("triage"))
	assert.NotEqual(t, "hub-sa-triage", ClientIdFor("triage"))
}

func TestFactoryStartsFromTheKeycloakClient(t *testing.T) {
	a := Factory(
		"org-uuid", "support-triage", "triages support mail", "hub-agent-support-triage",
		"kc-uuid", "user-uuid", "", "creator")

	assert.NotEmpty(t, a.Id)
	// The user id comes from Keycloak, not from hub: it is the client's own
	// service account user, and it is what makes the agent a principal the rest
	// of hub already understands.
	assert.Equal(t, "user-uuid", a.UserId)
	assert.Equal(t, "kc-uuid", a.KeycloakId)
	assert.Equal(t, "hub-agent-support-triage", a.ClientId)
	assert.Equal(t, "org-uuid", a.OrgId)
	assert.Equal(t, "creator", a.CreatedByUserId)
	// A new registration holds a freshly issued secret, and the stamp is the
	// only thing hub can ever say about that secret's age.
	assert.False(t, a.SecretRotatedAt.IsZero())
	// Only one method is issued today; the field exists so that moving an agent
	// off a shared secret does not need a flag day.
	assert.Equal(t, AuthMethodClientSecret, a.AuthMethod)
	assert.True(t, a.Root())
}

func TestFactoryRecordsLineage(t *testing.T) {
	a := Factory("org-uuid", "sub", "", "hub-agent-sub", "kc", "user", "parent-uuid", "creator")

	assert.Equal(t, "parent-uuid", a.ParentAgentId)
	assert.False(t, a.Root())
}
