package utils

import (
	"encoding/json"
	"fmt"
	"main/globals"
	"main/vulnerabilities/sql-injection/dialects"
	"net"
	"strings"
)

func KeyExists[K comparable, V any](m map[K]V, key K) bool {
	_, exists := m[key]
	return exists
}

func KeyMustExist[K comparable, V any](m map[K]V, key K) {
	if _, exists := m[key]; !exists {
		panic(fmt.Sprintf("Key %v does not exist in map!", key))
	}
}

func GetFromMap[T any](m map[string]interface{}, key string) *T {
	value, ok := m[key]
	if !ok {
		return nil
	}
	result, ok := value.(T)
	if !ok {
		return nil
	}
	return &result
}

func MustGetFromMap[T any](m map[string]interface{}, key string) T {
	value := GetFromMap[T](m, key)
	if value == nil {
		panic(fmt.Sprintf("Error parsing JSON: key %s does not exist or it has an incorrect type", key))
	}
	return *value
}

func ParseFormData(data string, separator string) map[string]interface{} {
	result := map[string]interface{}{}
	parts := strings.Split(data, separator)
	for _, part := range parts {
		keyValue := strings.Split(part, "=")
		if len(keyValue) != 2 {
			continue
		}
		result[keyValue[0]] = keyValue[1]
	}
	return result
}

func ParseBody(body string) map[string]interface{} {
	// first we check if the body is a string, and if it is, we try to parse it as JSON
	// if it fails, we parse it as form data

	if strings.HasPrefix(body, "{") && strings.HasSuffix(body, "}") {
		// if the body is a JSON object, we parse it as JSON
		jsonBody := map[string]interface{}{}
		err := json.Unmarshal([]byte(body), &jsonBody)
		if err == nil {
			return jsonBody
		}
	}
	return ParseFormData(body, "&")
}

func ParseQuery(query string) map[string]interface{} {
	return ParseFormData(query, "&")
}

func ParseCookies(cookies string) map[string]interface{} {
	return ParseFormData(cookies, ";")
}

func ParseHeaders(headers string) map[string]interface{} {
	j := map[string]interface{}{}
	err := json.Unmarshal([]byte(headers), &j)
	if err != nil {
		return map[string]interface{}{}
	}
	return j
}

func isIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func isLocalhost(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsLoopback()
}

func IsIpAllowed(allowedIps map[string]bool, ip string) bool {
	if globals.EnvironmentConfig.LocalhostAllowedByDefault && isLocalhost(ip) {
		return true
	}

	if len(allowedIps) == 0 {
		// No IPs configured in the allow list -> no restrictions
		return true
	}

	if KeyExists(allowedIps, ip) {
		return true
	}

	return false
}

func IsIpBypassed(ip string) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if globals.EnvironmentConfig.LocalhostAllowedByDefault && isLocalhost(ip) {
		return true
	}

	if KeyExists(globals.CloudConfig.BypassedIps, ip) {
		return true
	}

	return false
}

func getIpFromXForwardedFor(value string) string {
	forwardedIps := strings.Split(value, ",")
	for _, ip := range forwardedIps {
		ip = strings.TrimSpace(ip)
		if strings.Contains(ip, ":") {
			parts := strings.Split(ip, ":")
			if len(parts) == 2 {
				ip = parts[0]
			}
		}
		if isIP(ip) {
			return ip
		}
	}
	return ""
}

func GetIpFromRequest(remoteAddress string, xForwardedFor string) string {
	if xForwardedFor != "" && globals.EnvironmentConfig.TrustProxy {
		ip := getIpFromXForwardedFor(xForwardedFor)
		if isIP(ip) {
			return ip
		}
	}

	if remoteAddress != "" && isIP(remoteAddress) {
		return remoteAddress
	}

	return ""
}

func GetBlockingMode() int {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	return globals.CloudConfig.Block
}

func IsBlockingEnabled() bool {
	return GetBlockingMode() == 1
}

func IsUserBlocked(userID string) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	return KeyExists(globals.CloudConfig.BlockedUserIds, userID)
}

func ArrayContains(array []string, search string) bool {
	for _, member := range array {
		if member == search {
			return true
		}
	}
	return false
}

func GetSqlDialectFromString(dialect string) dialects.SQLDialect {
	dialect = strings.ToLower(dialect)
	switch dialect {
	case "mysql":
		return dialects.SQLDialectMySQL{}
	case "sqlite":
		return dialects.SQLDialectSQLite{}
	case "postgres":
		return dialects.SQLDialectPostgres{}
	default:
		return dialects.SQLDialectMySQL{} // default -> MySQL
	}
}

// StringPointer is a helper function to return a pointer to a string value.
func StringPointer(s string) *string {
	return &s
}

func BoolPointer(b bool) *bool {
	return &b
}
