package webscanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckQuery(t *testing.T) {
	t.Run("It detects injection patterns", func(t *testing.T) {
		testStrings := []string{
			"' or '1'='1",
			"1: SELECT * FROM users WHERE '1'='1'",
			"', information_schema.tables",
			"1' sleep(5)",
			"WAITFOR DELAY 1",
			"../etc/passwd",
		}
		for _, test := range testStrings {
			assert.Equal(t, true, checkQuery(map[string]interface{}{"test": test}))
		}
	})
	t.Run("It does not detect", func(t *testing.T) {
		nonMatchingQueryElements := []string{"google.de", "some-string", "1", ""}
		for _, test := range nonMatchingQueryElements {
			assert.Equal(t, false, checkQuery(map[string]interface{}{"test": test}))
		}
	})
	t.Run("It handles empty query object", func(t *testing.T) {
		assert.Equal(t, false, checkQuery(map[string]interface{}{}))
	})
	t.Run("It handles non-string query parameters", func(t *testing.T) {
		assert.Equal(t, false, checkQuery(map[string]interface{}{"test": 123}))
	})
}
