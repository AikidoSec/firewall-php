package utils

import (
	"fmt"
	"main/helpers"
	"main/log"
	"net"
	"net/netip"
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

func DecodeURIComponent(input string) string {
	var result strings.Builder
	length := len(input)

	for i := 0; i < length; {
		char := input[i]

		if char == '+' {
			result.WriteByte(' ')
			i++
			continue
		}

		if char == '%' && i+2 < length {
			first := decodeHexChar(input[i+1])
			second := decodeHexChar(input[i+2])

			if first != -1 && second != -1 {
				result.WriteByte(byte((first << 4) | second))
				i += 3
				continue
			}
		}

		result.WriteByte(char)
		i++
	}

	return result.String()
}

func decodeHexChar(ch byte) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	default:
		return -1
	}
}

func ParseFormData(data string, separator string) map[string]interface{} {
	result := map[string]interface{}{}
	parts := strings.Split(data, separator)
	for _, part := range parts {
		index := strings.Index(part, "=")
		if index == -1 {
			continue
		}

		key := part[:index]
		value := part[index+1:]

		// See: https://github.com/php/php-src/blob/master/main/php_variables.c#L313. PHP ignores duplicate cookie names per rfc2965
		// If the user supplies 2 cookies with the same key, we should not overwrite it to ensure our parsing is similar to PHP
		// Form and query parameters could potentially support ; as a separator (via PHP config `arg_separator.input`).
		// We assume that the user of this function wants to parse a cookie however.
		if separator == ";" && KeyExists(result, key) {
			continue
		}

		decodedValue := DecodeURIComponent(value)
		result[key] = decodedValue
	}
	return result
}

func ParseBody(body string) map[string]interface{} {
	// first we check if the body is a string, and if it is, we try to parse it as JSON
	// if it fails, we parse it as form data
	trimmedBody := strings.TrimSpace(body)
	if strings.HasPrefix(trimmedBody, "[") || strings.HasPrefix(trimmedBody, "{") {
		var jsonBody interface{}
		err := helpers.ParseJSON([]byte(trimmedBody), &jsonBody)
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
	err := helpers.ParseJSON([]byte(query), &jsonQuery)
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
	err := helpers.ParseJSON([]byte(headers), &j)
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

func IsIpAllowedOnEndpoint(server *ServerData, allowedIps *netipx.IPSet, ip string) int {
	if server == nil {
		return NoConfig
	}
	if server.AikidoConfig.LocalhostAllowedByDefault && isLocalhost(ip) {
		return Found
	}

	return IsIpInSet(allowedIps, ip)
}

func IsIpBypassed(server *ServerData, ip string) bool {
	if server == nil {
		return false
	}
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	return IsIpInSet(server.CloudConfig.BypassedIps, ip) == Found
}

func getIpFromXForwardedFor(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	parts := strings.Split(value, ",")
	for i := range parts {
		ip := strings.TrimSpace(parts[i])

		// If it's already a valid IP (prevents splitting on ':' for IPv6)
		if isIP(ip) {
			parts[i] = ip
			continue
		}

		// Normalize bracketed IPv6 without port: "[2001:db8::1]" -> "2001:db8::1"
		if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
			ip = ip[1 : len(ip)-1]
			parts[i] = ip
			continue
		}

		// IPv6 with port: "[2001:db8::1]:443" -> "2001:db8::1"
		// IPv4 with port: "203.0.113.5:1234" -> "203.0.113.5"
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
			parts[i] = ip
			continue
		}

		// Leave as-is; will validate below
		parts[i] = ip
	}

	// Pick the first valid, non-private IP
	for _, cand := range parts {
		if !isIP(cand) {
			continue
		}
		if !helpers.IsPrivateIP(cand) {
			return cand
		}
	}
	return ""
}

func GetIpFromRequest(server *ServerData, remoteAddress string, xForwardedFor string) string {
	if server == nil {
		return ""
	}
	if xForwardedFor != "" && server.AikidoConfig.TrustProxy {
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

func GetBlockingMode(server *ServerData) int {
	if server == nil {
		return NoConfig
	}
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	return server.CloudConfig.Block
}

func IsBlockingEnabled(server *ServerData) bool {
	return GetBlockingMode(server) == 1
}

func IsUserBlocked(server *ServerData, userID string) bool {
	if server == nil {
		return false
	}
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	return KeyExists(server.CloudConfig.BlockedUserIds, userID)
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

func IsIpAllowed(server *ServerData, ip string) bool {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	if helpers.IsPrivateIP(ip) {
		return true
	}

	result, _ := IsIpInList(server.CloudConfig.AllowedIps, ip)
	// IP is allowed if it's found in the allowed lists or if the allowed lists are not configured
	return result == Found || result == NoConfig
}

func IsIpBlocked(server *ServerData, ip string) (bool, []IpListMatch) {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	result, matches := IsIpInList(server.CloudConfig.BlockedIps, ip)
	return result == Found, matches
}

func IsIpMonitored(server *ServerData, ip string) (bool, []IpListMatch) {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	result, matches := IsIpInList(server.CloudConfig.MonitoredIps, ip)
	return result == Found, matches
}

func IsUserAgentInBlocklist(server *ServerData, userAgent string, blocklist *regexp.Regexp) (bool, []string) {
	if blocklist == nil {
		return false, []string{}
	}

	if blocklist.MatchString(userAgent) {
		matchedDetails := []string{}
		for key, valueRegex := range server.CloudConfig.UserAgentDetails {
			if valueRegex != nil && valueRegex.MatchString(userAgent) {
				matchedDetails = append(matchedDetails, key)
			}
		}

		return true, matchedDetails
	}

	return false, []string{}
}

func IsUserAgentBlocked(server *ServerData, userAgent string) (bool, []string) {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	return IsUserAgentInBlocklist(server, userAgent, server.CloudConfig.BlockedUserAgents)
}

func IsUserAgentMonitored(server *ServerData, userAgent string) (bool, []string) {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	return IsUserAgentInBlocklist(server, userAgent, server.CloudConfig.MonitoredUserAgents)
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

func AnonymizeToken(token string) string {
	if len(token) <= 4 {
		return "***" + token
	}
	return "***" + token[len(token)-4:]
}
