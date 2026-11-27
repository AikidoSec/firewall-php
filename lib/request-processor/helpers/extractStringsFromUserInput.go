package helpers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type PathPart struct {
	Type  string
	Key   string
	Index int
}

// buildPathToPayload builds the path to the payload
func buildPathToPayload(pathToPayload []PathPart) string {
	if len(pathToPayload) == 0 {
		return "."
	}

	path := ""
	for _, part := range pathToPayload {
		switch part.Type {
		case "jwt":
			path += "<jwt>"
		case "object":
			path += "." + part.Key
		case "array":
			path += fmt.Sprintf(".[%d]", part.Index)
		}
	}
	return path
}

const maxDepth = 1024

// extractStringsFromUserInput recursively extracts strings from user input and stops at maxDepth to prevent stack overflow
func ExtractStringsFromUserInput(obj interface{}, pathToPayload []PathPart, depth int) map[string]string {
	results := make(map[string]string)
	collectInputStrings(obj, buildPathToPayload(pathToPayload), depth, results, len(pathToPayload) == 0)
	return results
}

func collectInputStrings(obj interface{}, path string, depth int, results map[string]string, atRoot bool) {
	if depth > maxDepth {
		return
	}

	prefix := path
	if atRoot {
		prefix = ""
	}
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			keyStr := fmt.Sprintf("%v", iter.Key().Interface())
			results[keyStr] = path
			collectInputStrings(iter.Value().Interface(), prefix+"."+keyStr, depth+1, results, false)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			collectInputStrings(val.Index(i).Interface(), prefix+".["+strconv.Itoa(i)+"]", depth+1, results, false)
		}

		// Add array as string to results
		// This prevents bypassing the firewall by HTTP Parameter Pollution
		// Example: ?param=value1&param=value2 will be treated as array by express
		// If its used inside a string, it will be converted to a comma separated string
		if val.Len() > 0 {
			var values []string
			for i := 0; i < val.Len(); i++ {
				values = append(values, reflect.ValueOf(val.Index(i).Interface()).String())
			}
			results[strings.Join(values, ",")] = path
		}

	case reflect.String:
		str := val.String()
		results[str] = path
		jwt := tryDecodeAsJWT(str)
		if jwt.JWT {
			collectInputStrings(jwt.Object, prefix+"<jwt>", depth+1, results, false)
		}
	}
}

func ExtractResourceOrOriginal(filePath string) string {
	if len(filePath) >= len("php://filter/") && strings.ToLower(filePath[:len("php://filter/")]) == "php://filter/" {
		// https://github.com/php/php-src/blob/8b61c49987750b74bee19838c7f7c9fbbf53aace/ext/standard/php_fopen_wrapper.c#L348
		index := strings.LastIndex(filePath, "/resource=")
		if index != -1 {
			return filePath[index+len("/resource="):]
		}
	}
	return filePath
}
