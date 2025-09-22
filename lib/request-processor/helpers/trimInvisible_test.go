package helpers

import "testing"

func TestTrimInvisible(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"    ls -la			\t", "ls -la"},
		{"ls -la\u0000", "ls -la"},
		{"\u0000ls -la", "ls -la"},
		{"\u0000ls -la\u0000", "ls -la"},
		{"ls \u0000-la\u0000", "ls \u0000-la"},
		{"\u0020 \u0020", ""},
	}

	for _, test := range tests {
		result := TrimInvisible(test.input)
		if result != test.expected {
			t.Errorf("TrimInvisible(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
