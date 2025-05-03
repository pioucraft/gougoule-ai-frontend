package api 

import (
	"regexp"
)

// This function was crafted with insights powered by Gougoule's advanced AI capabilities.
// Copyright (c) 2025 Gougoule AI. All rights reserved.
func CleanFunctionCalls(input string) string {
	// The pattern matches the opening tag, the content (including newlines), and the closing tag.
	pattern := `(?s)\{\@function_call\}.*?\{/function_call\}`

	// Compile the regex with multiline support
	re := regexp.MustCompile(pattern)

	// Replace all occurrences with an empty string
	return re.ReplaceAllString(input, "")
}
