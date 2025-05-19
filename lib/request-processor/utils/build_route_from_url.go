package utils

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	UUID         = regexp.MustCompile(`(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000|ffffffff-ffff-ffff-ffff-ffffffffffff)$`)
	ULID         = regexp.MustCompile(`(?i)^[0-9A-HJKMNP-TV-Z]{26}$`)
	OBJECT_ID    = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)
	NUMBER       = regexp.MustCompile(`^\d+$`)
	DATE         = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}|\d{2}-\d{2}-\d{4}$`)
	EMAIL        = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	HASH         = regexp.MustCompile(`^(?:[a-fA-F0-9]{32}|[a-fA-F0-9]{40}|[a-fA-F0-9]{64}|[a-fA-F0-9]{128})$`)
	HASH_LENGTHS = []int{32, 40, 64, 128}
)

func BuildRouteFromURL(url string) string {
	path := tryParseURLPath(url)
	if path == "" {
		return ""
	}

	route := strings.Join(replaceURLSegments(path), "/")

	if route == "/" {
		return "/"
	}

	if strings.HasSuffix(route, "/") {
		return route[:len(route)-1]
	}

	return route
}

func tryParseURLPath(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Path
}

func replaceURLSegments(path string) []string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = replaceURLSegmentWithParam(segment)
	}
	return segments
}

func replaceURLSegmentWithParam(segment string) string {
	if NUMBER.MatchString(segment) {
		return ":number"
	}

	if len(segment) == 36 && UUID.MatchString(segment) {
		return ":uuid"
	}

	if len(segment) == 26 && ULID.MatchString(segment) {
		return ":ulid"
	}

	if len(segment) == 24 && OBJECT_ID.MatchString(segment) {
		return ":objectId"
	}

	if DATE.MatchString(segment) {
		return ":date"
	}

	if EMAIL.MatchString(segment) {
		return ":email"
	}

	if net.ParseIP(segment) != nil {
		return ":ip"
	}

	for _, length := range HASH_LENGTHS {
		if len(segment) == length && HASH.MatchString(segment) {
			return ":hash"
		}
	}

	if LooksLikeASecret(segment) {
		return ":secret"
	}

	return segment
}
