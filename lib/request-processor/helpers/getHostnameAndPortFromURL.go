package helpers

import (
	"strconv"
	"strings"
)

func ParsePort(portStr string) uint32 {
	if portStr == "" {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	if port < 0 {
		return 0
	}
	return uint32(port)
}

/*
The PHP libraries for outgoing requests (ex: curl), can take an invalid url ('http:/example.com')
and fix it themselves before making the actual request.
We need to replicate the same process as we get the exact URL that was written in the PHP script.
*/
func FixURL(url string) string {
	if !strings.HasPrefix(url, "https://") && strings.HasPrefix(url, "https:/") {
		return strings.Replace(url, "https:/", "https://", 1)
	}
	if !strings.HasPrefix(url, "http://") && strings.HasPrefix(url, "http:/") {
		return strings.Replace(url, "http:/", "http://", 1)
	}
	return url
}

func GetHostnameAndPortFromURL(url string) (string, uint32) {
	parsedURL := TryParseURL(FixURL(url))
	if parsedURL == nil {
		return "", 0
	}
	return parsedURL.Hostname(), GetPortFromURL(parsedURL)
}
