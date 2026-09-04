package organization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKindAppliesEverywhere(t *testing.T) {
	assert.True(t, KindPlatform.AppliesEverywhere(),
		"the operator's grants have to reach every tenant, or introducing organizations revokes every administrator")
	assert.False(t, KindBusiness.AppliesEverywhere())
	assert.False(t, KindPersonal.AppliesEverywhere(),
		"an individual is a tenant of one, not an operator")
}

func TestValidSlug(t *testing.T) {
	tests := []struct {
		slug string
		want bool
	}{
		{"acme", true},
		{"acme-corp", true},
		{"abc", true},
		{"a1b", true},
		{"", false},
		{"a", false},
		// Three characters is the floor, as it is for a service account name:
		// a slug this short is almost always a typo for a real one.
		{"a1", false},
		{"-acme", false},
		{"acme-", false},
		{"Acme", false},
		{"acme corp", false},
		{"acme/corp", false},
		{"acme.corp", false},
	}
	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidSlug(tt.slug))
		})
	}
}

func TestFactoryStartsActive(t *testing.T) {
	o := Factory("Acme", "acme", KindBusiness, "a customer")

	assert.NotEmpty(t, o.Id)
	assert.Equal(t, Active, o.Status)
	assert.Equal(t, KindBusiness, o.Kind)
	assert.False(t, o.Platform())
	assert.False(t, o.CreatedAt.IsZero())
}

func TestPlatformOrganization(t *testing.T) {
	o := &Organization{Id: PlatformOrgId, Kind: KindPlatform}
	assert.True(t, o.Platform())
}
