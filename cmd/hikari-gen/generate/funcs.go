package generate

import (
	"strings"
	"unicode"
)

// camelcase converts strings like "light-multi_zone" or "123-mode.enabled" to Go-style PascalCase.
// It removes delimiters like '-', '_', '.', and whitespace, and capitalizes each word.
// If the result starts with a digit, it prefixes it with an underscore.
func camelcase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '.' || r == '_' || unicode.IsSpace(r)
	})

	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}

	result := strings.Join(parts, "")

	if len(result) > 0 && unicode.IsDigit(rune(result[0])) {
		result = "_" + result
	}

	return result
}
