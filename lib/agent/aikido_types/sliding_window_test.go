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

func TestAddSample(t *testing.T) {
	t.Run("adds sample to empty sliding window", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)

		assert.Equal(t, 1, len(sw.Samples))
		assert.Equal(t, "GET", sw.Samples[0].Method)
		assert.Equal(t, "/api/users", sw.Samples[0].Url)
	})

	t.Run("adds multiple different samples", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)
		sw.AddSample("POST", "/api/login", 15)
		sw.AddSample("DELETE", "/api/users/123", 15)

		assert.Equal(t, 3, len(sw.Samples))
		assert.Equal(t, "GET", sw.Samples[0].Method)
		assert.Equal(t, "/api/users", sw.Samples[0].Url)
		assert.Equal(t, "POST", sw.Samples[1].Method)
		assert.Equal(t, "/api/login", sw.Samples[1].Url)
		assert.Equal(t, "DELETE", sw.Samples[2].Method)
		assert.Equal(t, "/api/users/123", sw.Samples[2].Url)
	})

	t.Run("prevents duplicate samples with same method and URL", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)
		sw.AddSample("GET", "/api/users", 15) // duplicate
		sw.AddSample("GET", "/api/users", 15) // duplicate

		assert.Equal(t, 1, len(sw.Samples))
		assert.Equal(t, "GET", sw.Samples[0].Method)
		assert.Equal(t, "/api/users", sw.Samples[0].Url)
	})

	t.Run("allows same URL with different methods", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)
		sw.AddSample("POST", "/api/users", 15)
		sw.AddSample("DELETE", "/api/users", 15)

		assert.Equal(t, 3, len(sw.Samples))
		assert.Equal(t, "GET", sw.Samples[0].Method)
		assert.Equal(t, "POST", sw.Samples[1].Method)
		assert.Equal(t, "DELETE", sw.Samples[2].Method)
	})

	t.Run("allows same method with different URLs", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)
		sw.AddSample("GET", "/api/posts", 15)
		sw.AddSample("GET", "/api/comments", 15)

		assert.Equal(t, 3, len(sw.Samples))
		assert.Equal(t, "/api/users", sw.Samples[0].Url)
		assert.Equal(t, "/api/posts", sw.Samples[1].Url)
		assert.Equal(t, "/api/comments", sw.Samples[2].Url)
	})

	t.Run("enforces maximum of 10 samples", func(t *testing.T) {
		sw := NewSlidingWindow()

		// Add 12 unique samples
		for i := 0; i < 12; i++ {
			sw.AddSample("GET", "/api/endpoint"+string(rune('0'+i)), 10)
		}

		assert.Equal(t, 10, len(sw.Samples))
	})

	t.Run("does not add 11th sample even if unique", func(t *testing.T) {
		sw := NewSlidingWindow()

		// Add exactly 10 samples
		for i := 0; i < 10; i++ {
			sw.AddSample("GET", "/api/endpoint"+string(rune('0'+i)), 10)
		}
		assert.Equal(t, 10, len(sw.Samples))

		// Try to add an 11th unique sample
		sw.AddSample("POST", "/api/new-endpoint", 10)
		assert.Equal(t, 10, len(sw.Samples))

		// Verify the 11th sample was not added
		found := false
		for _, sample := range sw.Samples {
			if sample.Method == "POST" && sample.Url == "/api/new-endpoint" {
				found = true
				break
			}
		}
		assert.False(t, found)
	})

	t.Run("duplicates do not count toward 10 sample limit", func(t *testing.T) {
		sw := NewSlidingWindow()

		// Add 5 unique samples
		for i := 0; i < 5; i++ {
			sw.AddSample("GET", "/api/endpoint"+string(rune('0'+i)), 15)
		}

		// Try to add duplicates
		sw.AddSample("GET", "/api/endpoint0", 15) // duplicate
		sw.AddSample("GET", "/api/endpoint1", 15) // duplicate
		sw.AddSample("GET", "/api/endpoint2", 15) // duplicate

		assert.Equal(t, 5, len(sw.Samples))

		// Add 5 more unique samples to reach 10
		for i := 5; i < 10; i++ {
			sw.AddSample("GET", "/api/endpoint"+string(rune('0'+i)), 15)
		}
		assert.Equal(t, 10, len(sw.Samples))

		// Try to add more duplicates - should still be 10
		sw.AddSample("GET", "/api/endpoint5", 15)
		sw.AddSample("GET", "/api/endpoint9", 15)
		assert.Equal(t, 10, len(sw.Samples))
	})

	t.Run("preserves samples during window operations", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("GET", "/api/users", 15)
		sw.AddSample("POST", "/api/login", 15)
		sw.Increment()

		// Advance the window
		sw.Advance(5)

		// Samples should still be present
		assert.Equal(t, 2, len(sw.Samples))
		assert.Equal(t, "GET", sw.Samples[0].Method)
		assert.Equal(t, "/api/users", sw.Samples[0].Url)
		assert.Equal(t, "POST", sw.Samples[1].Method)
		assert.Equal(t, "/api/login", sw.Samples[1].Url)
	})

	t.Run("empty method and URL are valid samples", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.AddSample("", "", 15)
		sw.AddSample("GET", "", 15)
		sw.AddSample("", "/api/users", 15)

		assert.Equal(t, 3, len(sw.Samples))
	})
}

func TestSlidingWindowSamplesIntegration(t *testing.T) {
	t.Run("simulates attack wave detection with samples", func(t *testing.T) {
		// Create a map simulating per-IP tracking
		ipMap := map[string]*SlidingWindow{
			"192.168.1.100": NewSlidingWindow(),
		}

		ip := "192.168.1.100"
		sw := ipMap[ip]

		// Simulate suspicious requests
		requests := []struct {
			method string
			url    string
		}{
			{"GET", "/admin"},
			{"GET", "/admin"}, // duplicate
			{"POST", "/admin"},
			{"GET", "/wp-admin"},
			{"GET", "/.env"},
			{"GET", "/config.php"},
			{"POST", "/login"},
			{"GET", "/admin"}, // duplicate
		}

		for _, req := range requests {
			sw.Increment()
			sw.AddSample(req.method, req.url, 15)
		}

		// Should have 6 unique samples (2 duplicates removed)
		assert.Equal(t, 6, len(sw.Samples))
		assert.Equal(t, 8, sw.Total) // But total count should be 8

		// Verify samples are unique
		uniqueCheck := make(map[string]bool)
		for _, sample := range sw.Samples {
			key := sample.Method + ":" + sample.Url
			assert.False(t, uniqueCheck[key], "Found duplicate sample: "+key)
			uniqueCheck[key] = true
		}
	})

	t.Run("samples persist across window advances until removal", func(t *testing.T) {
		windowMap := map[string]*SlidingWindow{
			"10.0.0.1": NewSlidingWindow(),
		}

		sw := windowMap["10.0.0.1"]
		sw.AddSample("GET", "/api/v1/users", 15)
		sw.AddSample("POST", "/api/v1/login", 15)
		sw.Increment()
		sw.Increment()

		// Advance window multiple times
		AdvanceSlidingWindowMap(windowMap, 3)
		AdvanceSlidingWindowMap(windowMap, 3)

		// Samples should still be there
		assert.Contains(t, windowMap, "10.0.0.1")
		assert.Equal(t, 2, len(windowMap["10.0.0.1"].Samples))

		// Advance until window is empty
		AdvanceSlidingWindowMap(windowMap, 3)

		// Window should be removed when total reaches 0
		assert.NotContains(t, windowMap, "10.0.0.1")
	})
}
