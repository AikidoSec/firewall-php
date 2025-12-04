package utils

import (
	"errors"
	"main/globals"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	UUID                    = regexp.MustCompile(`(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000|ffffffff-ffff-ffff-ffff-ffffffffffff)$`)
	ULID                    = regexp.MustCompile(`(?i)^[0-9A-HJKMNP-TV-Z]{26}$`)
	OBJECT_ID               = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)
	NUMBER                  = regexp.MustCompile(`^\d+$`)
	DATE                    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}|\d{2}-\d{2}-\d{4}$`)
	EMAIL                   = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	HASH                    = regexp.MustCompile(`^(?:[a-fA-F0-9]{32}|[a-fA-F0-9]{40}|[a-fA-F0-9]{64}|[a-fA-F0-9]{128})$`)
	HASH_LENGTHS            = []int{32, 40, 64, 128}
	PARAM_NAME_REGEXP       = regexp.MustCompile(`^[a-zA-Z_]+$`)
	MAX_REPLACEMENTS_NUMBER = 4
	PLACEHOLDER_REGEXP      = regexp.MustCompile(`\{[a-zA-Z_]+\}`)
)

func IsValidParamName(param string) bool {
	return PARAM_NAME_REGEXP.MatchString(param)
}

func CompileCustomPattern(pattern string) (*regexp.Regexp, error) {
	if !strings.Contains(pattern, "{") || !strings.Contains(pattern, "}") {
		return nil, errors.New("pattern should contain { or }")
	}

	if strings.Contains(pattern, "/") {
		return nil, errors.New("pattern should not contain slashes")
	}

	supported := map[string]string{
		"{digits}": `\d+`,
		"{alpha}":  "[a-zA-Z]+",
	}

	for name := range supported {
		if strings.Contains(pattern, name+name) {
			return nil, errors.New("pattern should not contain consecutive similar placeholders")
		}
	}

	var regexParts []string
	replacementsNumber := 0
	lastIndex := 0

	for _, match := range PLACEHOLDER_REGEXP.FindAllStringIndex(pattern, -1) {
		if match[0] > lastIndex {
			regexParts = append(regexParts, regexp.QuoteMeta(pattern[lastIndex:match[0]]))
		}

		placeholder := pattern[match[0]:match[1]]
		if replacement, ok := supported[placeholder]; ok {
			regexParts = append(regexParts, replacement)
			replacementsNumber++
			if replacementsNumber > MAX_REPLACEMENTS_NUMBER {
				return nil, errors.New("too many replacements in pattern")
			}
		} else {
			regexParts = append(regexParts, regexp.QuoteMeta(placeholder))
		}

		lastIndex = match[1]
	}

	// Add any remaining literal text after the last placeholder
	if lastIndex < len(pattern) {
		regexParts = append(regexParts, regexp.QuoteMeta(pattern[lastIndex:]))
	}

	compiled, err := regexp.Compile("^" + strings.Join(regexParts, "") + "$")
	if err != nil {
		return nil, err
	}

	return compiled, nil
}

func BuildRouteFromURL(url string) string {
	path := tryParseURLPath(url)
	if path == "" {
		return ""
	}

	route := strings.Join(replaceURLSegments(path), "/")

	if route == "/" || route == "" {
		return "/"
	}

	return strings.TrimRight(route, "/")
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
	newSegments := make([]string, 0, len(segments))
	for i, segment := range segments {
		if segment == "" && i != 0 {
			continue
		}
		newSegments = append(newSegments, replaceURLSegmentWithParam(segment))
	}
	return newSegments
}

func replaceURLSegmentWithParam(segment string) string {
	server := globals.GetCurrentServer()
	if server != nil {
		paramMatchers := server.ParamMatchers
		for param, regex := range paramMatchers {
			if regex.MatchString(segment) {
				return ":" + param
			}
		}
	}

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
