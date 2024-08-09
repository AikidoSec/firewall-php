package helpers

import (
	"encoding/json"
	"strings"
	"testing"
)

func jsonEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}
func TestTryDecodeAsJWT(t *testing.T) {
	t.Run("it returns false for empty string", func(t *testing.T) {
		result := tryDecodeAsJWT("")
		expected := JWTDecodeResult{JWT: false}
		if result != expected {
			t.Errorf("got %v, want %v", result, expected)
		}
	})

	t.Run("it returns false for invalid JWT", func(t *testing.T) {
		invalidJWTs := []string{"invalid", "invalid.invalid", "invalid.invalid.invalid", "invalid.invalid.invalid.invalid"}
		expected := JWTDecodeResult{JWT: false}
		for _, jwt := range invalidJWTs {
			result := tryDecodeAsJWT(jwt)
			if result != expected {
				t.Errorf("got %v, want %v", result, expected)
			}
		}
	})

	t.Run("it returns payload for invalid JWT", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected JWTDecodeResult
		}{
			{"/;ping%20localhost;.e30=.", JWTDecodeResult{JWT: true, Object: map[string]interface{}{}}},
			{"/;ping%20localhost;.W10=.", JWTDecodeResult{JWT: true, Object: []interface{}{}}},
		}

		for _, testCase := range testCases {
			result := tryDecodeAsJWT(testCase.input)
			if !jsonEqual(result, testCase.expected) {
				t.Errorf("got %v, want %v", result, testCase.expected)
			}
		}
	})

	t.Run("it returns the decoded JWT for valid JWT", func(t *testing.T) {
		input := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcm5hbWUiOnsiJG5lIjpudWxsfSwiaWF0IjoxNTE2MjM5MDIyfQ._jhGJw9WzB6gHKPSozTFHDo9NOHs3CNOlvJ8rWy6VrQ"
		expected := JWTDecodeResult{
			JWT: true,
			Object: map[string]interface{}{
				"sub": "1234567890",
				"username": map[string]interface{}{
					"$ne": nil,
				},
				"iat": 1516239022,
			},
		}
		result := tryDecodeAsJWT(input)
		if !jsonEqual(result, expected) {
			t.Errorf("got %v, want %v", result, expected)
		}
	})

	t.Run("it returns the decoded JWT for valid JWT with bearer prefix", func(t *testing.T) {
		input := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcm5hbWUiOnsiJG5lIjpudWxsfSwiaWF0IjoxNTE2MjM5MDIyfQ._jhGJw9WzB6gHKPSozTFHDo9NOHs3CNOlvJ8rWy6VrQ"
		expected := JWTDecodeResult{
			JWT: true,
			Object: map[string]interface{}{
				"sub": "1234567890",
				"username": map[string]interface{}{
					"$ne": nil,
				},
				"iat": 1516239022,
			},
		}
		result := tryDecodeAsJWT(strings.TrimPrefix(input, "Bearer "))
		if !jsonEqual(result, expected) {
			t.Errorf("got %v, want %v", result, expected)
		}
	})
}
