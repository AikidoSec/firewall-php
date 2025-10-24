package webscanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebScanMethod(t *testing.T) {
	t.Run("Test isWebScanMethod", func(t *testing.T) {
		assert.Equal(t, true, isWebScanMethod("BADMETHOD"))
		assert.Equal(t, true, isWebScanMethod("BADHTTPMETHOD"))
		assert.Equal(t, true, isWebScanMethod("BADDATA"))
		assert.Equal(t, true, isWebScanMethod("BADMTHD"))
		assert.Equal(t, true, isWebScanMethod("BDMTHD"))
	})
	t.Run("Test is not a web scan method", func(t *testing.T) {
		assert.Equal(t, false, isWebScanMethod("GET"))
		assert.Equal(t, false, isWebScanMethod("POST"))
		assert.Equal(t, false, isWebScanMethod("PUT"))
		assert.Equal(t, false, isWebScanMethod("DELETE"))
		assert.Equal(t, false, isWebScanMethod("PATCH"))
		assert.Equal(t, false, isWebScanMethod("OPTIONS"))
		assert.Equal(t, false, isWebScanMethod("HEAD"))
		assert.Equal(t, false, isWebScanMethod("TRACE"))
		assert.Equal(t, false, isWebScanMethod("CONNECT"))
		assert.Equal(t, false, isWebScanMethod("PURGE"))
	})
}
