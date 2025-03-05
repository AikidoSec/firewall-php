package utils

import (
	"encoding/json"
	"fmt"
	"main/globals"
	"main/log"
	"net"
	"runtime"
	"strings"

	"github.com/seancfoley/ipaddress-go/ipaddr"
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

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// unescape, ishex and unhex was forked from https://cs.opensource.google/go/go/+/refs/tags/go1.24.1:src/net/url/url.go;l=277.
// License: https://cs.opensource.google/go/go/+/refs/tags/go1.24.1:LICENSE.
// Copyright 2009 The Go Authors.
// What was changed: adjusted to work in Go's encodeQueryComponent mode only, and never throw errors
func unescape(s string) string {
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				i++
				continue
			}
			i += 3
		case '+':
			hasPlus = true
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s
	}

	var t strings.Builder
	t.Grow(len(s) - 2*n)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			// ensure we preserve the original characters if invalid
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				t.WriteByte(s[i])
				continue
			}
			t.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 2
		case '+':
			t.WriteByte(' ')
		default:
			t.WriteByte(s[i])
		}
	}
	return t.String()
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

		decodedValue := unescape(keyValue[1])
		if decodedValue != keyValue[1] {
			result[keyValue[0]] = decodedValue
		}
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

func IsIpBlocked(ip string) (bool, string) {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	ipAddress, err := ipaddr.NewIPAddressString(ip).ToAddress()
	if err != nil {
		log.Infof("Invalid ip address: %s\n", ip)
		return false, ""
	}

	for _, ipBlocklist := range globals.CloudConfig.BlockedIps {
		if (ipAddress.IsIPv4() && ipBlocklist.TrieV4.ElementContains(ipAddress.ToIPv4())) ||
			(ipAddress.IsIPv6() && ipBlocklist.TrieV6.ElementContains(ipAddress.ToIPv6())) {
			return true, ipBlocklist.Description
		}
	}

	return false, ""
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
