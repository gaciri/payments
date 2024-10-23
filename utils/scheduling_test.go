package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	testCases := []struct {
		attempts int
		expected time.Duration
	}{
		{attempts: 0, expected: 1 * time.Second},  // 2^0 * 1s = 1s
		{attempts: 1, expected: 2 * time.Second},  // 2^1 * 1s = 2s
		{attempts: 2, expected: 4 * time.Second},  // 2^2 * 1s = 4s
		{attempts: 3, expected: 8 * time.Second},  // 2^3 * 1s = 8s
		{attempts: 4, expected: 16 * time.Second}, // 2^4 * 1s = 16s
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Attempt %d", tc.attempts), func(t *testing.T) {
			result := ExponentialBackoff(tc.attempts)
			assert.Equal(t, tc.expected, result)
		})
	}
}
