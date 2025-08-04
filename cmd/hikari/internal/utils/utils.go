package utils

import (
	"cmp"
	"fmt"
	"slices"
)

const (
	DefaultPadding = 5
)

// RightPadder returns consistent right padding for a string in a list.
// If width is not supplied it uses the default width.
func RightPadder[S ~[]E, E any](ss S, lenFunc func(E) int, width ...int) func(s string) string {
	longest := slices.MaxFunc(ss, func(a, b E) int {
		return cmp.Compare(lenFunc(a), lenFunc(b))
	})

	longestLength := lenFunc(longest)
	maxWidth := longestLength + DefaultPadding
	if len(width) > 0 && width[0] >= longestLength {
		maxWidth = width[0]
	}

	return func(s string) string {
		return fmt.Sprintf("%-*s", maxWidth, s)
	}
}
