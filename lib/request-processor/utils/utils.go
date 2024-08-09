package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func CheckIfKeyExists[K comparable, V any](m map[K]V, key K) {
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

func FixURL(url string) string {
	if !strings.HasPrefix(url, "https://") && strings.HasPrefix(url, "https:/") {
		return strings.Replace(url, "https:/", "https://", 1)
	}
	if !strings.HasPrefix(url, "http://") && strings.HasPrefix(url, "http:/") {
		return strings.Replace(url, "http:/", "http://", 1)
	}
	return url
}

func GetDomain(rawurl string) string {
	parsedURL, err := url.Parse(FixURL(rawurl))
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
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

func ParseContext(context map[string]interface{}) map[string]interface{} {
	if context == nil {
		return map[string]interface{}{}
	}
	// first we check if the body is a string, and if it is, we try to parse it as JSON
	// if it fails, we parse it as form data
	if body, ok := context["body"].(string); ok {
		if strings.HasPrefix(body, "{") && strings.HasSuffix(body, "}") {
			// if the body is a JSON object, we parse it as JSON
			jsonBody := map[string]interface{}{}
			err := json.Unmarshal([]byte(body), &jsonBody)
			if err == nil {
				context["body"] = jsonBody
			}
		} else {
			context["body"] = ParseFormData(body, "&")
		}
	}

	// for the query, we always parse it as form data
	if query, ok := context["query"].(string); ok {
		context["query"] = ParseFormData(query, "&")
	}

	// for cookies
	if cookies, ok := context["cookies"].(string); ok {
		context["cookies"] = ParseFormData(cookies, ";")
	}

	return context
}
