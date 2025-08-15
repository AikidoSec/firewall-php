package helpers

import (
	"regexp"
	"strings"
)

func removeCTLByte(urlStr string) string {
	for i := 0; i < len(urlStr); i++ {
		if urlStr[i] < ' ' || urlStr[i] == 0x7f {
			urlStr = urlStr[:i] + urlStr[i+1:]
		}
	}
	return urlStr
}

var backslashAt = regexp.MustCompile(`\\+@`)

// If the urlStr contains \@ we need to replace it with @
// because the URL.Parse will fail to parse the url (invalid userinfo)
// IMPORTANT: there can be multiple backslashes before the @
func removeBackslashAt(urlStr string) string {
	return backslashAt.ReplaceAllString(urlStr, "@")
}

func removeWhitespace(urlStr string) string {
	return strings.ReplaceAll(urlStr, " ", "")
}

func removeUserInfo(raw string) string {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd == -1 {
		// No scheme, can't safely identify authority
		return raw
	}

	scheme := raw[:schemeEnd+3]
	rest := raw[schemeEnd+3:]

	// Authority is up to first '/', '?', or '#'
	authorityEnd := len(rest)
	for _, sep := range []string{"/", "?", "#"} {
		if idx := strings.Index(rest, sep); idx != -1 && idx < authorityEnd {
			authorityEnd = idx
		}
	}

	authority := rest[:authorityEnd]
	path := rest[authorityEnd:]

	// Remove userinfo if present (use LAST @)
	if at := strings.LastIndex(authority, "@"); at != -1 {
		authority = authority[at+1:]
	}

	return scheme + authority + path
}

func NormalizeRawUrl(urlStr string) string {
	urlStr = removeUserInfo(urlStr)
	urlStr = removeCTLByte(urlStr)
	urlStr = removeBackslashAt(urlStr)
	urlStr = removeWhitespace(urlStr)
	return urlStr
}
