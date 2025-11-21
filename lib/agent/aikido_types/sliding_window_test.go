package aikido_types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSlidingWindow(t *testing.T) {
	t.Run("initializes with correct default values", func(t *testing.T) {
		sw := NewSlidingWindow()
		assert.NotNil(t, sw)
		assert.Equal(t, 0, sw.Total)
		assert.Equal(t, 1, sw.Queue.Length())
	})

	t.Run("first bucket is initialized to zero", func(t *testing.T) {
		sw := NewSlidingWindow()
		assert.Equal(t, 0, sw.Queue.Get(0))
	})
}

func TestSlidingWindowIncrement(t *testing.T) {
	t.Run("increments total and current bucket", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment()
		assert.Equal(t, 1, sw.Total)
		assert.Equal(t, 1, sw.Queue.Get(0))
	})

	t.Run("multiple increments work correctly", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment()
		sw.Increment()
		sw.Increment()
		assert.Equal(t, 3, sw.Total)
		assert.Equal(t, 3, sw.Queue.Get(0))
	})

	t.Run("handles empty queue by creating a bucket", func(t *testing.T) {
		sw := &SlidingWindow{
			Queue: NewQueue[int](0),
		}
		sw.Increment()
		assert.Equal(t, 1, sw.Total)
		assert.Equal(t, 1, sw.Queue.Length())
	})
}

func TestSlidingWindowAdvance(t *testing.T) {
	t.Run("adds new bucket within window size", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment()
		sw.Increment()
		initialTotal := sw.Total

		sw.Advance(5)
		assert.Equal(t, 2, sw.Queue.Length())
		assert.Equal(t, initialTotal, sw.Total)
		assert.Equal(t, 0, sw.Queue.Get(1)) // new bucket is zero
	})

	t.Run("evicts oldest bucket when at capacity", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment() // bucket 0: 1
		sw.Advance(2)
		sw.Increment() // bucket 1: 1
		sw.Increment() // bucket 1: 2

		assert.Equal(t, 3, sw.Total)
		assert.Equal(t, 2, sw.Queue.Length())

		sw.Advance(2) // should evict bucket 0 (value: 1)
		assert.Equal(t, 2, sw.Total)
		assert.Equal(t, 2, sw.Queue.Length())
		assert.Equal(t, 2, sw.Queue.Get(0)) // old bucket 1 is now bucket 0
	})

	t.Run("handles multiple advances with eviction", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment() // bucket 0: 1
		sw.Advance(3)
		sw.Increment()
		sw.Increment() // bucket 1: 2
		sw.Advance(3)
		sw.Increment()
		sw.Increment()
		sw.Increment() // bucket 2: 3

		assert.Equal(t, 6, sw.Total)
		assert.Equal(t, 3, sw.Queue.Length())

		sw.Advance(3) // should evict bucket 0 (value: 1)
		assert.Equal(t, 5, sw.Total)
		assert.Equal(t, 3, sw.Queue.Length())

		sw.Advance(3) // should evict old bucket 1 (value: 2)
		assert.Equal(t, 3, sw.Total)
		assert.Equal(t, 3, sw.Queue.Length())
	})

	t.Run("prevents negative total with safety check", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Total = 5
		sw.Queue.Push(10) // bucket with value higher than total (shouldn't happen normally)

		sw.Advance(2)                // evict bucket with value 10, but total is only 5
		assert.Equal(t, 5, sw.Total) // should not go negative
	})
}

func TestSlidingWindowIsEmpty(t *testing.T) {
	t.Run("returns true for new sliding window", func(t *testing.T) {
		sw := NewSlidingWindow()
		assert.True(t, sw.IsEmpty())
	})

	t.Run("returns false after increment", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Increment()
		assert.False(t, sw.IsEmpty())
	})

	t.Run("returns true when total is zero", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Total = 0
		assert.True(t, sw.IsEmpty())
	})

	t.Run("returns false when total is non-zero", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.Total = 10
		assert.False(t, sw.IsEmpty())
	})
}

func TestAdvanceSlidingWindowMap(t *testing.T) {
	t.Run("advances all windows in the map", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{
			"key1": NewSlidingWindow(),
			"key2": NewSlidingWindow(),
		}
		windowMap["key1"].Increment()
		windowMap["key2"].Increment()
		windowMap["key2"].Increment()

		AdvanceSlidingWindowMap(windowMap, 5)

		assert.Equal(t, 2, windowMap["key1"].Queue.Length())
		assert.Equal(t, 2, windowMap["key2"].Queue.Length())
	})

	t.Run("removes empty windows from map", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{
			"key1": NewSlidingWindow(),
			"key2": NewSlidingWindow(),
			"key3": NewSlidingWindow(),
		}
		windowMap["key1"].Increment() // has count
		// key2 and key3 remain at zero

		AdvanceSlidingWindowMap(windowMap, 5)

		assert.Contains(t, windowMap, "key1")
		assert.NotContains(t, windowMap, "key2")
		assert.NotContains(t, windowMap, "key3")
	})

	t.Run("handles eviction and removal together", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{
			"key1": NewSlidingWindow(),
			"key2": NewSlidingWindow(),
		}

		// key1: add counts that will be evicted
		windowMap["key1"].Increment() // bucket 0: 1
		windowMap["key1"].Advance(2)
		windowMap["key1"].Increment() // bucket 1: 1

		// key2: remains empty
		windowMap["key2"].Advance(2)

		// Now advance with window size 2
		// key1: bucket 0 (value: 1) should be evicted, leaving total=1
		// key2: total is 0, should be removed
		AdvanceSlidingWindowMap(windowMap, 2)

		assert.Contains(t, windowMap, "key1")
		assert.NotContains(t, windowMap, "key2")
		assert.Equal(t, 1, windowMap["key1"].Total)
	})

	t.Run("handles empty map", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{}
		AdvanceSlidingWindowMap(windowMap, 5)
		assert.Equal(t, 0, len(windowMap))
	})

}

func TestSlidingWindowIntegration(t *testing.T) {
	t.Run("simulates typical usage pattern", func(t *testing.T) {
		sw := NewSlidingWindow()
		windowSize := 5

		// Time bucket 0
		sw.Increment()
		sw.Increment()
		assert.Equal(t, 2, sw.Total)

		// Time bucket 1
		sw.Advance(windowSize)
		sw.Increment()
		assert.Equal(t, 3, sw.Total)

		// Time bucket 2
		sw.Advance(windowSize)
		sw.Increment()
		sw.Increment()
		sw.Increment()
		assert.Equal(t, 6, sw.Total)

		// Time bucket 3
		sw.Advance(windowSize)
		assert.Equal(t, 6, sw.Total)

		// Time bucket 4
		sw.Advance(windowSize)
		sw.Increment()
		assert.Equal(t, 7, sw.Total)

		// Time bucket 5 - should evict bucket 0 (2 counts)
		sw.Advance(windowSize)
		assert.Equal(t, 5, sw.Total)

		// Time bucket 6 - should evict bucket 1 (1 count)
		sw.Advance(windowSize)
		assert.Equal(t, 4, sw.Total)

		// Time bucket 7 - should evict bucket 2 (3 counts)
		sw.Advance(windowSize)
		assert.Equal(t, 1, sw.Total)
	})

	t.Run("map with multiple windows over time", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{}
		windowSize := 3

		// Initialize windows with different patterns
		for _, key := range []string{"endpoint1", "endpoint2", "endpoint3"} {
			windowMap[key] = NewSlidingWindow()
		}

		// Time bucket 0
		windowMap["endpoint1"].Increment()
		windowMap["endpoint2"].Increment()

		// Time bucket 1
		AdvanceSlidingWindowMap(windowMap, windowSize)
		windowMap["endpoint1"].Increment()
		// endpoint3 stays at 0, will be removed

		// Verify state
		assert.Equal(t, 2, windowMap["endpoint1"].Total)
		assert.Equal(t, 1, windowMap["endpoint2"].Total)
		assert.NotContains(t, windowMap, "endpoint3")
	})
}
