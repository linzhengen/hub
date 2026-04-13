package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hello world",
			input:    "hello world",
			expected: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MD5String(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSHA1String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hello world",
			input:    "hello world",
			expected: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SHA1String(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
