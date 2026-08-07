package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/linzhengen/hub/v1/server/internal/usecase/pagination"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		limit          uint32
		offset         uint32
		expectedLimit  uint
		expectedOffset uint
	}{
		{
			// The whole point of the type: an omitted limit used to mean
			// "every row", which is what made a full table scan the default.
			name:          "no limit asked for falls back to the default",
			limit:         0,
			expectedLimit: uint(pagination.DefaultLimit),
		},
		{
			name:           "a limit within the bounds is kept",
			limit:          10,
			offset:         20,
			expectedLimit:  10,
			expectedOffset: 20,
		},
		{
			name:          "the maximum is allowed",
			limit:         pagination.MaxLimit,
			expectedLimit: uint(pagination.MaxLimit),
		},
		{
			name:          "a limit beyond the maximum is capped rather than rejected",
			limit:         pagination.MaxLimit + 1,
			expectedLimit: uint(pagination.MaxLimit),
		},
		{
			name:          "an absurd limit is capped too",
			limit:         1 << 31,
			expectedLimit: uint(pagination.MaxLimit),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := pagination.New(tt.limit, tt.offset)

			assert.Equal(t, tt.expectedLimit, page.Limit())
			assert.Equal(t, tt.expectedOffset, page.Offset())
		})
	}
}

func TestBoundsAreSane(t *testing.T) {
	assert.Positive(t, pagination.DefaultLimit)
	assert.LessOrEqual(t, pagination.DefaultLimit, pagination.MaxLimit,
		"the default must be reachable without being capped")
}
