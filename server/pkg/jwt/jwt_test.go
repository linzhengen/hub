package jwt

import (
	"testing"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestStringField(t *testing.T) {
	claims := jwtgo.MapClaims{
		"string_field": "hello",
		"int_field":    123,
	}

	tests := []struct {
		name      string
		fieldName string
		expected  string
	}{
		{
			name:      "existing string field",
			fieldName: "string_field",
			expected:  "hello",
		},
		{
			name:      "non-existing field",
			fieldName: "not_exist",
			expected:  "",
		},
		{
			name:      "non-string field",
			fieldName: "int_field",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := StringField(claims, tt.fieldName)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFloat64Field(t *testing.T) {
	claims := jwtgo.MapClaims{
		"float_field":  123.45,
		"string_field": "hello",
	}

	tests := []struct {
		name      string
		fieldName string
		expected  float64
	}{
		{
			name:      "existing float64 field",
			fieldName: "float_field",
			expected:  123.45,
		},
		{
			name:      "non-existing field",
			fieldName: "not_exist",
			expected:  0,
		},
		{
			name:      "non-float64 field",
			fieldName: "string_field",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Float64Field(claims, tt.fieldName)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIsMember(t *testing.T) {
	claims := jwtgo.MapClaims{
		"scope1": []any{"group-a", "group-b"},
		"scope2": "group-c",
	}

	tests := []struct {
		name     string
		groups   []string
		scopes   []string
		expected bool
	}{
		{
			name:     "is member",
			groups:   []string{"group-c", "group-d"},
			scopes:   []string{"scope1", "scope2"},
			expected: true,
		},
		{
			name:     "is not member",
			groups:   []string{"group-d", "group-e"},
			scopes:   []string{"scope1", "scope2"},
			expected: false,
		},
		{
			name:     "no matching scopes",
			groups:   []string{"group-a"},
			scopes:   []string{"scope3"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsMember(&claims, tt.groups, tt.scopes)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid token",
			token:    "a.b.c",
			expected: true,
		},
		{
			name:     "invalid token - too few parts",
			token:    "a.b",
			expected: false,
		},
		{
			name:     "invalid token - too many parts",
			token:    "a.b.c.d",
			expected: true, // Note: IsValid only checks for 3 parts, not exact format.
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValid(tt.token)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
