package aikido_types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQueue(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		maxSize := 2
		q := NewQueue[string](maxSize)
		assert.Nil(t, q.PushAndGetRemovedItemIfMaxExceeded("a"))
		assert.Nil(t, q.PushAndGetRemovedItemIfMaxExceeded("b"))
		expected := "a"
		removedItem := q.PushAndGetRemovedItemIfMaxExceeded("c")
		assert.NotNil(t, removedItem)
		assert.Equal(t, expected, *removedItem)
		assert.Equal(t, maxSize, q.Length())
		assert.False(t, q.IsEmpty())
	})

	t.Run("it can clear the queue", func(t *testing.T) {
		q := NewQueue[string](2)
		q.PushAndGetRemovedItemIfMaxExceeded("a")
		q.Clear()
		assert.Equal(t, 0, q.Length())
		assert.True(t, q.IsEmpty())
	})
}
