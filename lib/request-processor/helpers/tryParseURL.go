package helpers

import (
	"net"
	"net/url"

	"golang.org/x/net/idna"
)

func TryParseURL(input string) *url.URL {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil
	}

	// Convert the host to Unicode if it's an IDN (https://www.rfc-editor.org/rfc/rfc3492)
	parsedHost, err := idna.ToUnicode(parsedURL.Host)
	if err == nil {
		parsedURL.Host = parsedHost
	}

	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err == nil {
		ip := net.ParseIP(host)
		if ip != nil {
			parsedURL.Host = ip.String() + ":" + port
		} else {
			parsedURL.Host = host + ":" + port
		}
	}

	return parsedURL
}
