package helpers

import (
	"net/url"
	"strconv"
)

func GetPortFromURL(u *url.URL) uint32 {
	// If a port is explicitly specified and it's a valid integer
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 0 {
			return 0
		}
		return uint32(port)
	}

	// Default ports based on protocol
	switch u.Scheme {
	case "https":
		return 443
	case "http":
		return 80
	}
	return 0
}
