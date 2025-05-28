package aikido_types

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitingQueue(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		q := RateLimitingQueue{}
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

	t.Run("pop returns -1 if queue is empty", func(t *testing.T) {
		q := RateLimitingQueue{}
		assert.Equal(t, -1, q.Pop())
	})

	t.Run("increment last checks if queue is empty", func(t *testing.T) {
		q := RateLimitingQueue{}
		q.IncrementLast()
		assert.Equal(t, 0, q.Length())
	})
} 