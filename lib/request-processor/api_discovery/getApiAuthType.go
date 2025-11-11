package api_discovery

import (
	"main/context"
	"main/ipc/protos"
	"slices"
	"strings"
)

// Common API key header and cookie names
var commonApiKeyHeaderNames = []string{
	"x-api-key",
	"api-key",
	"apikey",
	"x-token",
	"token",
}

var commonAuthCookieNames = append([]string{
	"auth",
	"session",
	"jwt",
	"token",
	"sid",
	"connect.sid",
	"auth_token",
	"access_token",
	"refresh_token",
}, commonApiKeyHeaderNames...)

// GetApiAuthType returns the authentication type of the API request.
// Returns nil if the authentication type could not be determined.
func GetApiAuthType() []*protos.APIAuthType {
	var result []*protos.APIAuthType

	headers := context.GetHeadersParsed()

	// Check the Authorization header
	authHeader, authHeaderExists := headers["authorization"].(string)
	if authHeaderExists {
		authHeaderType := getAuthorizationHeaderType(authHeader)
		if authHeaderType != nil {
			result = append(result, authHeaderType)
		}
	}

	result = append(result, findApiKeys()...)
	return result
}

// getAuthorizationHeaderType returns the authentication type from the Authorization header.
func getAuthorizationHeaderType(authHeader string) *protos.APIAuthType {
	if len(authHeader) == 0 {
		return nil
	}
	if strings.Contains(authHeader, " ") {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 {
			authType := parts[0]
			if isHTTPAuthScheme(authType) {
				scheme := strings.ToLower(authType)
				return &protos.APIAuthType{
					Type:   "http",
					Scheme: scheme,
				}
			}
		}
	}

	// Default to apiKey if the auth type is not recognized
	name := "Authorization"
	return &protos.APIAuthType{
		Type: "apiKey",
		In:   ("header"),
		Name: name,
	}
}

func getPhpHttpHeaderEquivalent(apiKey string) string {
	return strings.ReplaceAll(apiKey, "-", "_")
}

// findApiKeys searches for API keys in headers and cookies.
func findApiKeys() []*protos.APIAuthType {
	var result []*protos.APIAuthType

	headers := context.GetHeadersParsed()
	cookies := context.GetCookiesParsed()
	for header_index, header := range commonApiKeyHeaderNames {
		if value, exists := headers[getPhpHttpHeaderEquivalent(header)]; exists && value != "" {
			result = append(result, &protos.APIAuthType{
				Type: "apiKey",
				In:   "header",
				Name: commonApiKeyHeaderNames[header_index],
			})
		}
	}

	if len(cookies) > 0 {
		for cookieName := range cookies {
			lowerCookieName := strings.ToLower(cookieName)
			if slices.Contains(commonAuthCookieNames, lowerCookieName) {
				result = append(result, &protos.APIAuthType{
					Type: "apiKey",
					In:   "cookie",
					Name: cookieName,
				})
			}
		}
	}

	return result
}

// isHTTPAuthScheme checks if the given string is a valid HTTP authentication scheme.
// You will need to implement this function similar to the TypeScript helper function.
func isHTTPAuthScheme(scheme string) bool {
	// You can add proper logic here to check the scheme, e.g., "basic", "bearer", etc.
	// For example:
	allowedSchemes := []string{"basic", "bearer"}
	return slices.Contains(allowedSchemes, strings.ToLower(scheme))
}
