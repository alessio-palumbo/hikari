package generate

import "testing"

func Test_camelcase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"light", "Light"},
		{"multi zone", "MultiZone"},
		{"light-multi_zone", "LightMultiZone"},
		{"123-mode.enabled", "_123ModeEnabled"},
		{"some-thing_here.now", "SomeThingHereNow"},
		{"", ""},
		{" alreadyCamelCase ", "AlreadyCamelCase"},
		{"UPPER_CASE", "UPPERCASE"},
		{"has spaces-and.dashes", "HasSpacesAndDashes"},
		{"snake_case_name", "SnakeCaseName"},
		{"a b c", "ABC"},
	}

	for _, test := range tests {
		result := camelcase(test.input)
		if result != test.expected {
			t.Errorf("Camelcase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
