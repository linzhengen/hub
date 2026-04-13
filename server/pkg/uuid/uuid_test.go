package uuid

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUUID(t *testing.T) {
	id, err := NewUUID()
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	assert.Equal(t, uuid.Version(7), id.Version())
}

func TestMustUUID(t *testing.T) {
	assert.NotPanics(t, func() {
		id := MustUUID()
		assert.NotEqual(t, uuid.Nil, id)
	})
}

func TestMustString(t *testing.T) {
	idStr := MustString()
	// Regex for UUID, should match any version's format.
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	assert.Regexp(t, uuidRegex, idStr)
}
