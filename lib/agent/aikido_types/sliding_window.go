package aikido_types

// SuspiciousRequest represents a suspicious request sample collected during attack wave detection
type SuspiciousRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

// SlidingWindow represents a time-based sliding window counter.
// It maintains a queue of counts per time bucket and a running total.
type SlidingWindow struct {
	Total int        // Running total of all counts in the window
	Queue Queue[int] // Queue of counts per time bucket
  Samples  []SuspiciousRequest // Sample requests collected for attack wave detection (max 10)
}

// NewSlidingWindow creates a new sliding window with the specified size.
func NewSlidingWindow() *SlidingWindow {
	sw := &SlidingWindow{
		Queue: NewQueue[int](0), // no max size, we handle it manually
	}
	// Ensure there is a current bucket
	sw.Queue.Push(0)
	return sw
}

// Advance pushes a new (zeroed) time bucket,
// evicting the oldest if we exceed the window size, and adjusting total accordingly.
func (sw *SlidingWindow) Advance(windowSize int) {
	// If we're at capacity, remove the oldest bucket first
	if sw.Queue.Length() >= windowSize {
		dropped := sw.Queue.Pop()
		if sw.Total >= dropped { // safety check to avoid negative total
			sw.Total -= dropped
		}
	}
	// Add a new bucket for the current time period
	sw.Queue.Push(0)
}

// Increment increments the current time bucket.
func (sw *SlidingWindow) Increment() {
	if sw.Queue.IsEmpty() {
		sw.Queue.Push(0)
	}
	sw.Queue.IncrementLast()
	sw.Total++
}

// AddSample adds a sample request to the sliding window for attack wave detection.
// It maintains a maximum of 10 unique samples (based on method and URL).
func (sw *SlidingWindow) AddSample(method, url string) {
	const maxSamples = 10

	// Check if this sample already exists
	for _, sample := range sw.Samples {
		if sample.Method == method && sample.URL == url {
			return // Already exists, skip
		}
	}

	// Add the sample if we haven't reached the limit
	if len(sw.Samples) < maxSamples {
		sw.Samples = append(sw.Samples, SuspiciousRequest{
			Method: method,
			URL:    url,
		})
	}
}

// IsEmpty returns true if the total count is zero.
func (sw *SlidingWindow) IsEmpty() bool {
	return sw.Total == 0
}

// AdvanceSlidingWindowMap advances all sliding windows in the map and removes entries where Total is 0.
func AdvanceSlidingWindowMap(windowMap map[string]*SlidingWindow, windowSize int) {
	for key, window := range windowMap {
		// Advance the sliding window for this entry
		window.Advance(windowSize)
		// if total is 0, remove the entry
		if window.IsEmpty() {
			delete(windowMap, key)
		}
	}
}
