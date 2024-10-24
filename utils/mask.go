package utils

import "strings"

func MaskString(input string) string {
	length := len(input)

	if length <= 4 {
		return "****"
	}

	maskedPortion := strings.Repeat("*", length-4)
	return maskedPortion + input[length-4:]
}
