package gen

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	validPkgRegex    = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	validStructRegex = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
)

// ValidatePackageName checks if a string is a valid Go package name.
func ValidatePackageName(name string) error {
	if validPkgRegex.MatchString(name) {
		return nil
	}
	suggested := SanitizePackageName(name)
	return fmt.Errorf("invalid package name '%s'. Go package names may only contain lowercase letters, digits, and underscores, and must start with a letter. Suggested alternative: '%s'", name, suggested)
}

// ValidateStructName checks if a string is a valid exported Go struct name.
func ValidateStructName(name string) error {
	if validStructRegex.MatchString(name) {
		return nil
	}
	suggested := SanitizeStructName(name)
	return fmt.Errorf("invalid struct name '%s'. Go exported struct names may only contain letters, digits, and underscores, and must start with an uppercase letter. Suggested alternative: '%s'", name, suggested)
}

// SanitizePackageName converts a string into a valid Go package name.
func SanitizePackageName(input string) string {
	input = strings.ToLower(input)
	
	// Strip all non-alphanumeric/underscore characters
	var sb strings.Builder
	for _, ch := range input {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		}
	}
	
	result := sb.String()
	
	// Strip leading digits or underscores to ensure it starts with a letter
	// Actually, the rules say it can start with a letter. Let's find the first letter.
	firstLetterIdx := -1
	for i, ch := range result {
		if ch >= 'a' && ch <= 'z' {
			firstLetterIdx = i
			break
		}
	}
	
	if firstLetterIdx != -1 && firstLetterIdx > 0 {
		result = result[firstLetterIdx:]
	} else if firstLetterIdx == -1 {
		// If there are no letters at all, fallback to a default
		result = "pkg" + result
	}
	
	if result == "" {
		return "pkg"
	}
	return result
}

// SanitizeStructName converts a string into a valid exported Go struct name.
func SanitizeStructName(input string) string {
	var parts []string
	// Split by anything that isn't alphanumeric
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	for _, word := range strings.FieldsFunc(input, f) {
		if len(word) > 0 {
			// Title case the word
			parts = append(parts, strings.ToUpper(word[:1])+word[1:])
		}
	}
	
	result := strings.Join(parts, "")
	
	// Ensure it starts with an uppercase letter
	if len(result) > 0 {
		first := rune(result[0])
		if !unicode.IsLetter(first) {
			result = "Type" + result
		} else {
			result = strings.ToUpper(string(first)) + result[1:]
		}
	}
	
	if result == "" {
		return "Struct"
	}
	
	return result
}
