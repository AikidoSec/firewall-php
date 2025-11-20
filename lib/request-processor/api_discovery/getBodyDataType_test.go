package api_discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBodyDataType(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]interface{}
		expected string
	}{
		{"JSON content_type", map[string]interface{}{"content_type": "application/json"}, "json"},
		{"API JSON content_type", map[string]interface{}{"content_type": "application/vnd.api+json"}, "json"},
		{"CSP report content_type", map[string]interface{}{"content_type": "application/csp-report"}, "json"},
		{"X JSON content_type", map[string]interface{}{"content_type": "application/x-json"}, "json"},
		{"JSON with charset", map[string]interface{}{"content_type": "application/json; charset=utf-8"}, "json"},
		{"JSON uppercase", map[string]interface{}{"content_type": "Application/JSON"}, "json"},
		{"JSON LD", map[string]interface{}{"content_type": "application/ld+json"}, "json"},
		{"JSON with whitespace", map[string]interface{}{"content_type": " application/json "}, "json"},
		{"Form-urlencoded content_type", map[string]interface{}{"content_type": "application/x-www-form-urlencoded"}, "form-urlencoded"},
		{"Multipart form-data content_type", map[string]interface{}{"content_type": "multipart/form-data"}, "form-data"},
		{"XML content_type", map[string]interface{}{"content_type": "text/xml"}, "xml"},
		{"XML with +xml suffix", map[string]interface{}{"content_type": "application/atom+xml"}, "xml"},
		{"HTML content_type", map[string]interface{}{"content_type": "text/html"}, ""},
		{"Nonexistent content_type", map[string]interface{}{"x-test": "abc"}, ""},
		{"Null input", nil, ""},
		{"Empty headers", map[string]interface{}{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBodyDataType(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}
