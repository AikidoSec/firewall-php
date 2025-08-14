package helpers

import "strings"

func removeCTLByte(urlStr string) string {
	for i := 0; i < len(urlStr); i++ {
		if urlStr[i] < ' ' || urlStr[i] == 0x7f {
			urlStr = urlStr[:i] + urlStr[i+1:]
		}
	}
	return urlStr
}

// If the urlStr contains \@ we need to replace it with @
// because the URL.Parse will fail to parse the url (invalid userinfo)
func removeBackslashAt(urlStr string) string {
	return strings.ReplaceAll(urlStr, "\\@", "@")
}

func removeWhitespace(urlStr string) string {
	return strings.ReplaceAll(urlStr, " ", "")
}

func NormalizeRawUrl(urlStr string) string {
	urlStr = removeCTLByte(urlStr)
	urlStr = removeBackslashAt(urlStr)
	urlStr = removeWhitespace(urlStr)
	return urlStr
}
