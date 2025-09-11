package helpers

import (
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

func NormalizeRawUrl(urlStr string) string {
	urlStr = removeCTLByte(urlStr)
	urlStr = FixURL(urlStr)
	urlStr = removeUserInfo(urlStr)
	return urlStr
}
