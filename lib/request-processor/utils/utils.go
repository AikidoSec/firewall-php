package utils

import (
	"encoding/json"
	"fmt"
	"main/globals"
	"main/log"
	"net"
	"net/netip"
	"net/url"
	"runtime"
	"strings"

	. "main/aikido_types"

	"go4.org/netipx"
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

		// See: https://github.com/php/php-src/blob/master/main/php_variables.c#L313. PHP ignores duplicate cookie names per rfc2965
		// If the user supplies 2 cookies with the same key, we should not overwrite it to ensure our parsing is similar to PHP
		// Form and query parameters could potentially support ; as a separator (via PHP config `arg_separator.input`).
		// We assume that the user of this function wants to parse a cookie however.
		if separator == ";" && KeyExists(result, keyValue[0]) {
			continue
		}

		result[keyValue[0]] = keyValue[1]
		decodedValue, err := url.QueryUnescape(keyValue[1])
		if err == nil && decodedValue != keyValue[1] {
			result[keyValue[0]] = decodedValue
		}
	}
	return result
}

func ParseBody(body string) map[string]interface{} {
	// first we check if the body is a string, and if it is, we try to parse it as JSON
	// if it fails, we parse it as form data
	trimmedBody := strings.TrimSpace(body)
	if strings.HasPrefix(trimmedBody, "[") || strings.HasPrefix(trimmedBody, "{") {
		var jsonBody interface{}
		err := json.Unmarshal([]byte(trimmedBody), &jsonBody)
		if err == nil {
			if array, ok := jsonBody.([]interface{}); ok {
				return map[string]interface{}{"array": array}
			}

			if jsonObject, ok := jsonBody.(map[string]interface{}); ok {
				return jsonObject
			}
		}
	}

	return ParseFormData(body, "&")
}

func ParseQuery(query string) map[string]interface{} {
	jsonQuery := map[string]interface{}{}
	err := json.Unmarshal([]byte(query), &jsonQuery)
	if err == nil {
		return jsonQuery
	}
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

func IsIpInSet(ipSet *netipx.IPSet, ip string, defaultVal bool) bool {
	if ipSet == nil || ipSet.Equal(&netipx.IPSet{}) {
		// No IPs configured in the list -> return default value
		return defaultVal
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		log.Infof("Invalid ip address: %s\n", ip)
		return defaultVal
	}

	if ipSet.Contains(ipAddress) {
		return true
	}

	return false
}

func IsIpAllowedOnEndpoint(allowedIps *netipx.IPSet, ip string) bool {
	if globals.EnvironmentConfig.LocalhostAllowedByDefault && isLocalhost(ip) {
		return true
	}

	return IsIpInSet(allowedIps, ip, true)
}

func IsIpBypassed(ip string) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	return IsIpInSet(globals.CloudConfig.BypassedIps, ip, false)
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

func IsIpInSetList(ipList map[string]IpList, ip string, defaultVal bool) (bool, string) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if len(ipList) == 0 {
		return defaultVal, ""
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		log.Infof("Invalid ip address: %s\n", ip)
		return defaultVal, ""
	}

	for _, l := range ipList {
		if l.IpSet.Contains(ipAddress) {
			return true, l.Description
		}
	}

	return false, ""
}

func IsIpBlocked(ip string) (bool, string) {
	return IsIpInSetList(globals.CloudConfig.BlockedIps, ip, false)
}

func IsIpAllowed(ip string) (bool, string) {
	return IsIpInSetList(globals.CloudConfig.AllowedIps, ip, true)
}

func IsUserAgentBlocked(userAgent string) (bool, string) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if globals.CloudConfig.BlockedUserAgents == nil {
		return false, ""
	}

	if globals.CloudConfig.BlockedUserAgents.MatchString(userAgent) {
		return true, "bot detection"
	}

	return false, ""
}

type DatabaseType int

const (
	Generic DatabaseType = iota
	Ansi
	BigQuery
	Clickhouse
	Databricks
	DuckDB
	Hive
	MSSQL
	MySQL
	PostgreSQL
	Redshift
	Snowflake
	SQLite
)

func GetSqlDialectFromString(dialect string) int {
	dialect = strings.ToLower(dialect)
	switch dialect {
	case "mysql":
		return int(MySQL)
	case "sqlite":
		return int(SQLite)
	case "postgres":
		return int(PostgreSQL)
	default:
		return int(Generic)
	}
}

// StringPointer is a helper function to return a pointer to a string value.
func StringPointer(s string) *string {
	return &s
}

func BoolPointer(b bool) *bool {
	return &b
}

func ArrayContains(array []string, search string) bool {
	for _, member := range array {
		if member == search {
			return true
		}
	}
	return false
}

func GetArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	}
	panic(fmt.Sprintf("Running on unsupported architecture \"%s\"!", runtime.GOARCH))
}
