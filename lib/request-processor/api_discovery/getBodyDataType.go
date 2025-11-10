package api_discovery

import (
	. "main/aikido_types"
	"strings"
)

// getBodyDataType tries to determine the type of the body data based on the content type header.
func getBodyDataType(headers map[string]interface{}) string {
	if headers == nil {
		return Undefined
	}

	contentType, exists := headers["content_type"].(string)
	if !exists {
		return Undefined
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))

	if isJSONContentType(contentType) {
		return JSON
	}

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return FormURLEncoded
	}

	if strings.HasPrefix(contentType, "multipart/form-data") {
		return FormData
	}

	if isXMLContentType(contentType) {
		return XML
	}

	return Undefined
}

var jsonContentTypes = []string{
	"application/json",
	"application/csp-report",
	"application/x-json",
}

func isJSONContentType(contentType string) bool {
	for _, jsonType := range jsonContentTypes {
		if strings.HasPrefix(contentType, jsonType) {
			return true
		}
	}

	return strings.Contains(contentType, "+json")
}

func isXMLContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/xml") ||
		strings.HasPrefix(contentType, "text/xml") ||
		strings.Contains(contentType, "+xml")
}
