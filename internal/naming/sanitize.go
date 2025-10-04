package naming

import (
	"regexp"
	"strings"
)

var (
	// Regex to replace disallowed characters
	// Allowed characters: A-Z, a-z, 0-9, ., _, -
	invalidCharsRegex = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

	// Replace consecutive hyphens with a single hyphen
	multiHyphenRegex = regexp.MustCompile(`-+`)
)

// Sanitize converts a branch name to a filesystem-safe directory name
// Example: "feature/new-ui" -> "feature-new-ui"
func Sanitize(branchName string) string {
	// Convert slashes to hyphens
	s := strings.ReplaceAll(branchName, "/", "-")

	// Convert disallowed characters to hyphens
	s = invalidCharsRegex.ReplaceAllString(s, "-")

	// Collapse consecutive hyphens
	s = multiHyphenRegex.ReplaceAllString(s, "-")

	// Trim leading and trailing hyphens
	s = strings.Trim(s, "-")

	// Limit path length to avoid filesystem issues (max 255 chars on most systems)
	const maxLength = 200 // Conservative limit for safety
	if len(s) > maxLength {
		s = s[:maxLength]
		// Re-trim in case we cut in the middle of trailing hyphens
		s = strings.TrimRight(s, "-")
	}

	return s
}

// SanitizeWithLowercase converts a branch name to lowercase and sanitizes it
func SanitizeWithLowercase(branchName string) string {
	return Sanitize(strings.ToLower(branchName))
}
