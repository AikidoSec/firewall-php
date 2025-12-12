package helpers

import (
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

// NormalizeHostname normalizes a hostname for consistent comparison.
// It trims invisible characters, decodes URL percent-encoding,
// converts Punycode to Unicode (IDN), and converts to lowercase.
// This prevents bypass attacks using Punycode encoding (e.g., "xn--mnchen-3ya.de" vs "münchen.de")
// or URL encoding (e.g., "ssrf-r%C3%A9directs.com" vs "ssrf-rédirects.com").
func NormalizeHostname(hostname string) string {
	trimmed := TrimInvisible(hostname)

	// Decode URL percent-encoding (e.g., %C3%A9 -> é)
	decoded, err := url.QueryUnescape(trimmed)
	if err != nil {
		decoded = trimmed
	}

	// Convert to lowercase first - idna.ToUnicode requires lowercase "xn--" prefix
	lowercased := strings.ToLower(decoded)

	// Convert Punycode (xn--...) to Unicode form for consistent comparison
	// e.g., "xn--mnchen-3ya.de" -> "münchen.de"
	unicodeHostname, err := idna.ToUnicode(lowercased)
	if err != nil {
		// If conversion fails, use the lowercased hostname as-is
		unicodeHostname = lowercased
	}

	return unicodeHostname
}
