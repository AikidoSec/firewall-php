package api_discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsObject(t *testing.T) {
	t.Run("Test isObject", func(t *testing.T) {
		assert.Equal(t, false, isObject(""))
		assert.Equal(t, false, isObject([]string{"1"}))
		assert.Equal(t, false, isObject(nil))

		assert.Equal(t, true, isObject(map[string]any{"1": "2"}))
		assert.Equal(t, true, isObject(map[string]string{"1": "2"}))
		assert.Equal(t, true, isObject(map[string]int{"1": 500}))
		assert.Equal(t, true, isObject(map[string][]string{"1": {"2"}}))
		assert.Equal(t, true, isObject(map[string][]any{"1": {"2"}}))
	})
}
