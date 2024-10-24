package utils

import "testing"

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234", "****"},           // Case: Input length exactly 4
		{"12", "****"},             // Case: Input length less than 4
		{"12345", "*2345"},         // Case: Input length greater than 4
		{"123456789", "*****6789"}, // Case: Input length greater than 4
		{"", "****"},               // Case: Empty input
		{"abcde", "*bcde"},         // Case: Alphabetical characters
	}

	for _, test := range tests {
		result := MaskString(test.input)
		if result != test.expected {
			t.Errorf("MaskString(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}
