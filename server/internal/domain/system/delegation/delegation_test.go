package delegation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func at(offset time.Duration) *time.Time {
	t := time.Now().Add(offset)
	return &t
}

// TestLive covers the two ways a delegation ends, which end it for different
// reasons: revocation is a write and an expiry is the clock. A caller never has
// to know which, so both are answered here rather than at each call site.
func TestLive(t *testing.T) {
	now := time.Now()

	for _, tt := range []struct {
		name string
		d    Delegation
		live bool
	}{
		{name: "no expiry and not revoked", d: Delegation{}, live: true},
		{name: "expires later", d: Delegation{ExpiresAt: at(time.Hour)}, live: true},
		{name: "already expired", d: Delegation{ExpiresAt: at(-time.Hour)}, live: false},
		{name: "revoked", d: Delegation{RevokedAt: at(-time.Minute)}, live: false},
		{
			name: "revoked wins over an expiry that has not come",
			d:    Delegation{ExpiresAt: at(time.Hour), RevokedAt: at(-time.Minute)},
			live: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.live, tt.d.Live(now))
		})
	}
}

// The boundary is the moment itself: a delegation that expires now is over, the
// same way an expired grant is dropped by the decision rather than served for
// one more request.
func TestLive_ExpiryIsExclusive(t *testing.T) {
	now := time.Now()
	d := Delegation{ExpiresAt: &now}

	assert.False(t, d.Live(now))
	assert.True(t, d.Live(now.Add(-time.Nanosecond)))
}

// Revoked answers whatever the clock says, because the record outlives the
// delegation: "this agent could act for me until Tuesday" is what an
// investigation reads afterwards.
func TestRevoked(t *testing.T) {
	assert.False(t, (&Delegation{}).Revoked())
	assert.True(t, (&Delegation{RevokedAt: at(-time.Minute)}).Revoked())
	assert.True(t, (&Delegation{RevokedAt: at(time.Hour)}).Revoked())
}

func TestFactory(t *testing.T) {
	expiry := at(time.Hour)
	d := Factory("agent", "principal", "principal", "org", "for the nightly report",
		[]string{"perm-a", "perm-b"}, 2, expiry)

	require.NotEmpty(t, d.Id)
	assert.Equal(t, "agent", d.AgentId)
	assert.Equal(t, "principal", d.PrincipalUserId)
	assert.Equal(t, "org", d.OrgId)
	assert.Equal(t, []string{"perm-a", "perm-b"}, d.PermissionIds)
	assert.Equal(t, uint32(2), d.MaxDepth)
	assert.Equal(t, expiry, d.ExpiresAt)
	// A new delegation is live and has never been revoked.
	assert.Nil(t, d.RevokedAt)
	assert.True(t, d.Live(time.Now()))
}
