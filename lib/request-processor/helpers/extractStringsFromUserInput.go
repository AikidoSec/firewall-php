package helpers

import (
	"fmt"
	"reflect"
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

// extractStringsFromUserInput recursively extracts strings from user input
func ExtractStringsFromUserInput(obj interface{}, pathToPayload []PathPart) map[string]string {
	results := make(map[string]string)

	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Map:
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			results[keyStr] = buildPathToPayload(pathToPayload)
			nestedResults := ExtractStringsFromUserInput(val.MapIndex(key).Interface(), append(pathToPayload, PathPart{Type: "object", Key: keyStr}))
			for k, v := range nestedResults {
				results[k] = v
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			nestedResults := ExtractStringsFromUserInput(val.Index(i).Interface(), append(pathToPayload, PathPart{Type: "array", Index: i}))
			for k, v := range nestedResults {
				results[k] = v
			}
		}

	case reflect.String:
		str := val.String()
		results[str] = buildPathToPayload(pathToPayload)
		jwt := tryDecodeAsJWT(str)
		if jwt.JWT {
			for k, v := range ExtractStringsFromUserInput(jwt.Object, append(pathToPayload, PathPart{Type: "jwt"})) {
				results[k] = v
			}
		}

	}

	return results
}
