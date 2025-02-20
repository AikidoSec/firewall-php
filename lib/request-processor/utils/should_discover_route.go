package utils

import (
	"path/filepath"
	"strings"
)

const (
	NOT_FOUND          = 404
	METHOD_NOT_ALLOWED = 405
	MOVED_PERMANENTLY  = 301
	FOUND              = 302
	SEE_OTHER          = 303
	TEMPORARY_REDIRECT = 307
	PERMANENT_REDIRECT = 308
)

var EXCLUDED_METHODS = []string{"OPTIONS", "HEAD"}
var IGNORE_EXTENSIONS = []string{"properties", "config", "webmanifest"}
var ALLOW_EXTENSIONS = []string{"html", "php", "php3", "php4", "php5", "phtml"}
var IGNORE_STRINGS = []string{"cgi-bin"}

func ShouldDiscoverRoute(statusCode int, route, method string) bool {
	if containsStr(EXCLUDED_METHODS, method) {
		return false
	}

	if statusCode < 200 || statusCode > 399 {
		return false
	}

	segments := strings.Split(route, "/")

	// e.g. /path/to/.file or /.directory/file
	for _, segment := range segments {
		if isDotFile(segment) {
			return false
		}

		if containsIgnoredString(segment) {
			return false
		}
		
		if !isAllowedExtension(segment) {
			return false
		}
	}

	return true
}

func isAllowedExtension(segment string) bool {
	extension := filepath.Ext(segment)

	if extension != "" && strings.HasPrefix(extension, ".") {
		extension = extension[1:]

		if containsStr(ALLOW_EXTENSIONS, extension) {
			return true
		}

		if len(extension) >= 2 && len(extension) <= 5 {
			return false
		}

		if containsStr(IGNORE_EXTENSIONS, extension) {
			return false
		}
	}

	return true
}

func isDotFile(segment string) bool {
	if segment == ".well-known" {
		return false
	}

	return strings.HasPrefix(segment, ".") && len(segment) > 1
}

func containsIgnoredString(segment string) bool {
	for _, str := range IGNORE_STRINGS {
		if strings.Contains(segment, str) {
			return true
		}
	}
	return false
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
