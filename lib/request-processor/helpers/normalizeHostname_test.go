package helpers

import "testing"

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ASCII hostname unchanged",
			input:    "example.com",
			expected: "example.com",
		},
		{
			name:     "Uppercase to lowercase",
			input:    "EXAMPLE.COM",
			expected: "example.com",
		},
		{
			name:     "Punycode to Unicode",
			input:    "xn--mnchen-3ya.de",
			expected: "münchen.de",
		},
		{
			name:     "Punycode to Unicode with percent encoding",
			input:    "ssrf-r%C3%A9directs.testssandbox.com",
			expected: "ssrf-rédirects.testssandbox.com",
		},
		{
			name:     "Unicode stays Unicode",
			input:    "münchen.de",
			expected: "münchen.de",
		},
		{
			name:     "Punycode uppercase to lowercase Unicode",
			input:    "XN--MNCHEN-3YA.DE",
			expected: "münchen.de",
		},
		{
			name:     "Russian Cyrillic Punycode to Unicode",
			input:    "xn--80adxhks.ru",
			expected: "москва.ru",
		},
		{
			name:     "Chinese Punycode to Unicode",
			input:    "xn--fiq228c.com",
			expected: "中文.com",
		},
		{
			name:     "Mixed case Punycode subdomain",
			input:    "XN--BSE-SNA.evil.com",
			expected: "böse.evil.com",
		},
		{
			name:     "Hostname with leading/trailing whitespace",
			input:    "  example.com  ",
			expected: "example.com",
		},
		{
			name:     "Hostname with leading zero-width space",
			input:    "\u200Bexample.com",
			expected: "example.com",
		},
		{
			name:     "Hostname with trailing zero-width space",
			input:    "example.com\u200B",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHostname(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHostname(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
