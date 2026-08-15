package serviceaccount

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The name becomes a Keycloak client id and a username, so it is validated
// narrowly here rather than escaped in three places later.
func TestValidName(t *testing.T) {
	valid := []string{"ci-deploy", "nightly-backup-2", "a1b", "cid"}
	for _, name := range valid {
		assert.True(t, ValidName(name), "%q should be accepted", name)
	}

	invalid := map[string]string{
		"ab":                        "two characters is too short to describe anything",
		"ci":                        "likewise",
		"-ci":                       "leading hyphen",
		"ci-":                       "trailing hyphen",
		"CI":                        "uppercase would not survive a client id",
		"ci deploy":                 "a space",
		"ci/deploy":                 "a path separator",
		"ci@deploy":                 "an address separator",
		strings.Repeat("a", 65):     "longer than the column",
		"":                          "empty",
		"service-account-hub-sa-ci": "long enough, but see below",
	}
	for name, why := range invalid {
		if name == "service-account-hub-sa-ci" {
			// This one is actually well formed; it is here to record that a
			// name colliding with a generated username is not rejected by
			// shape. The uniqueness of `name` and `client_id` is what stops two
			// accounts sharing an identity.
			assert.True(t, ValidName(name))
			continue
		}
		assert.False(t, ValidName(name), "%q should be refused: %s", name, why)
	}
}

// The derived identifiers are what tie hub's row, Keycloak's client and the
// machine's user together. They are derived rather than chosen so that two
// service accounts cannot claim the same client.
func TestDerivedIdentifiers(t *testing.T) {
	assert.Equal(t, "hub-sa-ci-deploy", ClientIdFor("ci-deploy"))
	assert.Equal(t, "service-account-hub-sa-ci-deploy", UsernameFor("ci-deploy"))

	// `users.email` is unique and not null and a machine has no mailbox, so the
	// address is synthesised. `.invalid` is reserved by RFC 2606, so nothing
	// ever tries to deliver to it.
	assert.Equal(t, "hub-sa-ci-deploy@service-account.invalid", EmailFor("ci-deploy"))
	assert.True(t, strings.HasSuffix(EmailFor("anything"), ".invalid"))

	// Two names never collide, which is what keeps the unique columns honest.
	assert.NotEqual(t, EmailFor("ci-deploy"), EmailFor("ci-deploy-2"))
	assert.NotEqual(t, ClientIdFor("ci-deploy"), ClientIdFor("ci-deploy-2"))
}

func TestFactoryStartsFromTheKeycloakClient(t *testing.T) {
	account := Factory("ci-deploy", "deploys main", "hub-sa-ci-deploy", "kc-uuid", "user-uuid", "creator")

	assert.NotEmpty(t, account.Id)
	// The user id comes from Keycloak, not from hub: it is the client's own
	// service account user, and it is what makes the machine a principal the
	// rest of hub already understands.
	assert.Equal(t, "user-uuid", account.UserId)
	assert.Equal(t, "kc-uuid", account.KeycloakId)
	assert.Equal(t, "hub-sa-ci-deploy", account.ClientId)
	assert.Equal(t, "creator", account.CreatedByUserId)
}
