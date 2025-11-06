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

func TestQueueIntOperations(t *testing.T) {
	t.Run("it works with push, pop, and increment", func(t *testing.T) {
		q := NewQueue[int](0)
		q.Push(1)
		q.Push(2)
		q.Push(3)
		q.IncrementLast()
		assert.Equal(t, 3, q.Length())
		assert.Equal(t, 1, q.Pop())
		assert.Equal(t, 2, q.Pop())
		// because it was incremented
		assert.Equal(t, 4, q.Pop())
	})

	t.Run("pop returns 0 if queue is empty", func(t *testing.T) {
		q := NewQueue[int](0)
		assert.Equal(t, 0, q.Pop())
	})

	t.Run("increment last checks if queue is empty", func(t *testing.T) {
		q := NewQueue[int](0)
		q.IncrementLast()
		assert.Equal(t, 0, q.Length())
	})
}
