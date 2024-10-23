package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoinUrlPaths(t *testing.T) {
	testCases := []struct {
		name      string
		base      string
		path      string
		expected  string
		expectErr bool
	}{
		{
			name:     "Base URL without trailing slash, relative path",
			base:     "https://example.com",
			path:     "path",
			expected: "https://example.com/path",
		},
		{
			name:      "Invalid base URL",
			base:      "://invalid-url",
			path:      "path",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := JoinUrlPaths(tc.base, tc.path)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
