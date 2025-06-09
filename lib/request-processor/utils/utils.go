package utils

import (
	"encoding/json"
	"fmt"
	"main/globals"
	"main/helpers"
	"main/log"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"runtime"
	"strings"

	. "main/aikido_types"

	"go4.org/netipx"
)

const (
	NoConfig = -1
	NotFound = 0
	Found    = 1
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

func IsIpInSet(ipSet *netipx.IPSet, ip string) int {
	if ipSet == nil || ipSet.Equal(&netipx.IPSet{}) {
		// No IPs configured in the list -> return default value
		return NoConfig
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		log.Infof("Invalid ip address: %s\n", ip)
		return NoConfig
	}

	if ipSet.Contains(ipAddress) {
		return Found
	}

	return NotFound
}

func IsIpAllowedOnEndpoint(allowedIps *netipx.IPSet, ip string) int {
	if globals.EnvironmentConfig.LocalhostAllowedByDefault && isLocalhost(ip) {
		return Found
	}

	return IsIpInSet(allowedIps, ip)
}

func IsIpBypassed(ip string) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	return IsIpInSet(globals.CloudConfig.BypassedIps, ip) == Found
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

type IpListMatch struct {
	Key         string
	Description string
}

func IsIpInList(ipList map[string]IpList, ip string) (int, []IpListMatch) {
	if len(ipList) == 0 {
		return NoConfig, []IpListMatch{}
	}

	ipAddress, err := netip.ParseAddr(ip)
	if err != nil {
		return NoConfig, []IpListMatch{}
	}

	matches := []IpListMatch{}
	for listKey, list := range ipList {
		if list.IpSet.Contains(ipAddress) {
			matches = append(matches, IpListMatch{Key: listKey, Description: list.Description})
		}
	}

	if len(matches) == 0 {
		return NotFound, matches
	}

	return Found, matches
}

func IsIpAllowed(ip string) bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if helpers.IsPrivateIP(ip) {
		return true
	}

	result, _ := IsIpInList(globals.CloudConfig.AllowedIps, ip)
	// IP is allowed if it's found in the allowed lists or if the allowed lists are not configured
	return result != NotFound
}

func IsIpBlocked(ip string) (bool, []IpListMatch) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	result, matches := IsIpInList(globals.CloudConfig.BlockedIps, ip)
	return result == Found, matches
}

func IsIpMonitored(ip string) (bool, []IpListMatch) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	result, matches := IsIpInList(globals.CloudConfig.MonitoredIps, ip)
	return result == Found, matches
}

func IsUserAgentInBlocklist(userAgent string, blocklist *regexp.Regexp) (bool, []string) {
	if blocklist == nil {
		return false, []string{}
	}

	if blocklist.MatchString(userAgent) {
		matchedDetails := []string{}
		for key, valueRegex := range globals.CloudConfig.UserAgentDetails {
			if valueRegex != nil && valueRegex.MatchString(userAgent) {
				matchedDetails = append(matchedDetails, key)
			}
		}

		return true, matchedDetails
	}

	return false, []string{}
}

func IsUserAgentBlocked(userAgent string) (bool, []string) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	return IsUserAgentInBlocklist(userAgent, globals.CloudConfig.BlockedUserAgents)
}

func IsUserAgentMonitored(userAgent string) (bool, []string) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()
	return IsUserAgentInBlocklist(userAgent, globals.CloudConfig.MonitoredUserAgents)
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

func IsWildcardEndpoint(method, route string) bool {
	return method == "*" || strings.Contains(route, "*")
}
