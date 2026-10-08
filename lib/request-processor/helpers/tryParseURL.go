package helpers

import (
	"net/url"

	"golang.org/x/net/idna"
)

func TryParseURL(input string) *url.URL {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		return nil
	}

	// Convert the hostname to Unicode if it's an IDN (https://www.rfc-editor.org/rfc/rfc3492)
	// We must convert only the hostname part (without port) because idna.ToUnicode
	// fails when a port is present (e.g., "xn--mnchen-3ya.de:8080")
	hostname := parsedURL.Hostname()
	unicodeHostname, err := idna.ToUnicode(hostname)
	if err == nil && unicodeHostname != hostname {
		// Reconstruct Host with the Unicode hostname and original port (if any)
		port := parsedURL.Port()
		if port != "" {
			parsedURL.Host = unicodeHostname + ":" + port
		} else {
			parsedURL.Host = unicodeHostname
		}
	}
	return parsedURL
}
