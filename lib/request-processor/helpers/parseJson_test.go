package helpers

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseJSON(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      string
		expectedError error
	}{
		{
			name:     "negative large exponent",
			input:    `{ "age": -1e+9999}`,
			expected: `{"age":-1e+9999}`,
		},
		{
			name:     "decimal with large exponent",
			input:    `{ "age": 0.4e0066999999999999999999999999999999999}`,
			expected: `{"age":0.4e0066999999999999999999999999999999999}`,
		},
		{
			name:     "positive decimal with large exponent",
			input:    `{ "age": 1.5e+9999}`,
			expected: `{"age":1.5e+9999}`,
		},
		{
			name:     "negative integer with large exponent",
			input:    `{ "age": -123123e10000}`,
			expected: `{"age":-123123e10000}`,
		},
		{
			name:          "negative integer with large exponent and trailing garbage",
			input:         `{ "age": -123123e10000}{}`,
			expected:      ``,
			expectedError: errors.New("unexpected extra JSON values"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			err := ParseJSON([]byte(tc.input), &result)
			if err != nil {
				if tc.expectedError.Error() != err.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				return
			}

			resultJSON, err := json.Marshal(result)
			if err != nil {
				t.Errorf("Failed to marshal result: %v", err)
			}

			if string(resultJSON) != tc.expected {
				t.Errorf("Expected JSON string %q, got %q", tc.expected, resultJSON)
			}
		})
	}
}
