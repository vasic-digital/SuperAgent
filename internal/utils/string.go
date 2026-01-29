package utils

// ReverseString reverses a string, correctly handling UTF-8 characters.
func ReverseString(s string) string {
	// Convert string to rune slice to handle multi-byte UTF-8 characters
	runes := []rune(s)

	// Reverse the rune slice in place
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
