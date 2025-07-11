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

	// Convert the host to Unicode if it's an IDN (https://www.rfc-editor.org/rfc/rfc3492)
	parsedHost, err := idna.ToUnicode(parsedURL.Host)
	if err == nil {
		parsedURL.Host = parsedHost
	}
	return parsedURL
}
