package api_discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrimitiveType(t *testing.T) {
	testCases := []struct {
		str      string
		expected bool
	}{
		{str: "string", expected: true},
		{str: "number", expected: true},
		{str: "boolean", expected: true},
		{str: "null", expected: true},
		{str: "array", expected: false},
		{str: "object", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			result := isPrimitiveType(tc.str)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOnlyContainsPrimitiveTypes(t *testing.T) {
	testCases := []struct {
		name     string
		types    []string
		expected bool
	}{
		{name: "all primitive", types: []string{"string", "number", "boolean"}, expected: true},
		{name: "contains object", types: []string{"number", "object"}, expected: false},
		{name: "contains array", types: []string{"string", "array"}, expected: false},
		{name: "both object and array", types: []string{"object", "array"}, expected: false},
		{name: "empty", types: []string{}, expected: true},
		{name: "single primitive", types: []string{"string"}, expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := onlyContainsPrimitiveTypes(tc.types)
			assert.Equal(t, tc.expected, result)
		})
	}
}
