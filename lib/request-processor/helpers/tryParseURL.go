package helpers

import (
	"net"
	"net/url"
	"strconv"

	"golang.org/x/net/idna"
)

func TryParseURL(input string) *url.URL {
	parsedURL, err := url.Parse(input)
	if err != nil || parsedURL.Host == "" {
		return nil
	}

	// Convert the host to Unicode if it's an IDN (https://www.rfc-editor.org/rfc/rfc3492)
	parsedHost, err := idna.ToUnicode(parsedURL.Host)
	if err == nil {
		parsedURL.Host = parsedHost
	}
	// If the port is not present, we need to add it based on the scheme
	if parsedURL.Port() == "" {
		port := 0
		switch parsedURL.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
		parsedURL.Host = parsedURL.Host + ":" + strconv.Itoa(int(port))
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
