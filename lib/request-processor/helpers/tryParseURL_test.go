package helpers

import (
	"net/url"
	"testing"
)

func TestTryParseURL_InvalidURL(t *testing.T) {
	input := "invalid"
	result := TryParseURL(input)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestTryParseURL_ValidURL(t *testing.T) {
	input := "https://example.com:443"
	expected, _ := url.Parse(input)
	result := TryParseURL(input)

	if result == nil || result.String() != expected.String() {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
