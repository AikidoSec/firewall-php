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

// \@ -> @
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
