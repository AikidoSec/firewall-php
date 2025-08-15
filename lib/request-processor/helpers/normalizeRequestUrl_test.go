package helpers

import "testing"

func TestNormalizeRequestUrl(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://localhost:4000", "http://localhost:4000"},
		{"http://localhost:4000 ", "http://localhost:4000"},
		{"http://localhost:4000" + "\x00", "http://localhost:4000"},
		{"http://\\@localhost:4000", "http://@localhost:4000"},
		{"http://127.1.1.1:4000\\\\\\@127.0.0.1:80/", "http://127.1.1.1:4000@127.0.0.1:80/"},
	}
	for _, test := range tests {
		result := NormalizeRawUrl(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected %v but got %v", test.input, test.expected, result)
		}
	}
}
