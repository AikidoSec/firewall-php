package helpers

import (
	"net/url"
	"regexp"
	"strings"
)

// remove all control characters (< 32) and 0x7f(DEL) + whitespace
func removeCTLByte(urlStr string) string {
	for i := 0; i < len(urlStr); i++ {
		if urlStr[i] <= ' ' || urlStr[i] == 0x7f {
			urlStr = urlStr[:i] + urlStr[i+1:]
		}
	}
	return urlStr
}

func removeUserInfo(raw string) string {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd == -1 {
		// No scheme, can't safely identify authority
		return raw
	}

	scheme := raw[:schemeEnd+3]
	rest := raw[schemeEnd+3:]

	// Authority is up to first '/', '?', or '#' (https://datatracker.ietf.org/doc/html/rfc3986#section-3.2)
	authorityEnd := len(rest)
	for _, sep := range []string{"/", "?", "#"} {
		if idx := strings.Index(rest, sep); idx != -1 && idx < authorityEnd {
			authorityEnd = idx
		}
	}

	authority := rest[:authorityEnd]
	path := rest[authorityEnd:]

	// Remove userinfo if present
	if at := strings.LastIndex(authority, "@"); at != -1 {
		authority = authority[at+1:]
	}

	return scheme + authority + path
}

func UnescapeUrl(urlStr string) string {
	unescapedUrl, err := url.QueryUnescape(urlStr)
	if err != nil {
		return urlStr
	}
	return unescapedUrl
}

// ConvertIPv6Mapped converts IPv6-mapped IPv4 (only if it contains ::ffff:)
// Example: "http://[::ffff:10.0.0.1]" -> "http://10.0.0.1"
func convertIPv6Mapped(input string) string {
	// Return immediately if not IPv6-mapped form
	if !strings.Contains(input, "::ffff:") {
		return input
	}

	// Extract URL scheme if present (http://, https://, etc.)
	scheme := ""
	if strings.Contains(input, "://") {
		parts := strings.SplitN(input, "://", 2)
		scheme = parts[0] + "://"
		input = parts[1]
	}

	// Strip brackets
	input = strings.TrimPrefix(input, "[")
	input = strings.TrimSuffix(input, "]")

	// Replace ::ffff:x.x.x.x -> x.x.x.x
	re := regexp.MustCompile(`::ffff:(\d+\.\d+\.\d+\.\d+)`)
	ip := re.ReplaceAllString(input, "$1")

	return scheme + ip
}

func NormalizeRawUrl(urlStr string) string {
	urlStr = UnescapeUrl(urlStr)
	urlStr = removeCTLByte(urlStr)
	urlStr = FixURL(urlStr)
	urlStr = removeUserInfo(urlStr)
	urlStr = convertIPv6Mapped(urlStr)
	return urlStr
}
