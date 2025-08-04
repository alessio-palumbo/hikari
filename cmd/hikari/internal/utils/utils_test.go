package utils

import "testing"

func TestRightPadder(t *testing.T) {
	testCases := map[string]struct {
		options    []string
		lenFunc    func(string) int
		maxWidth   []int
		wantPadded []string
	}{
		"with default width": {
			options: []string{"red", "orange", "green", "yellow", "cyan", "blue", "magenta", "purple"},
			lenFunc: func(o string) int { return len(o) },
			wantPadded: []string{
				"red         ",
				"orange      ",
				"green       ",
				"yellow      ",
				"cyan        ",
				"blue        ",
				"magenta     ",
				"purple      ",
			},
		},
		"with set width": {
			options:  []string{"red", "orange", "green", "yellow", "cyan", "blue", "magenta", "purple"},
			lenFunc:  func(o string) int { return len(o) },
			maxWidth: []int{20},
			wantPadded: []string{
				"red                 ",
				"orange              ",
				"green               ",
				"yellow              ",
				"cyan                ",
				"blue                ",
				"magenta             ",
				"purple              ",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			padFunc := RightPadder(tc.options, tc.lenFunc, tc.maxWidth...)
			for i, o := range tc.options {
				got := padFunc(o)
				if got != tc.wantPadded[i] {
					t.Errorf("Expected padded string does not match: got [%s], want [%s]", got, tc.wantPadded[i])
				}
			}
		})
	}
}
