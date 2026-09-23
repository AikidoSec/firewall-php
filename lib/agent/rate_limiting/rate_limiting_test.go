package rate_limiting

import (
	. "main/aikido_types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdvanceTracksNextReset(t *testing.T) {
	started := time.Unix(1000, 0)
	for _, tc := range []struct {
		name        string
		elapsed     time.Duration
		wantBuckets int
		wantReset   time.Duration
	}{
		{"before tick", 55 * time.Second, 1, time.Minute},
		{"at tick", time.Minute, 2, 2 * time.Minute},
		{"missed tick", 130 * time.Second, 3, 3 * time.Minute},
		{"expired window", 190 * time.Second, 0, 4 * time.Minute},
		{"long pause", 24*time.Hour + 10*time.Second, 0, 24*time.Hour + time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			counts := func() map[string]*SlidingWindow {
				window := NewSlidingWindow()
				window.Increment()
				return map[string]*SlidingWindow{"identity": window}
			}
			endpoint := &RateLimitingValue{
				Config:      RateLimitingConfig{MaxRequests: 1, WindowSizeInMinutes: 3},
				NextResetAt: started.Add(time.Minute),
				UserCounts:  counts(), IpCounts: counts(), RateLimitGroupCounts: counts(),
			}
			now := started.Add(tc.elapsed)
			endpoint.Mutex.Lock()
			AdvanceEndpointQueues(endpoint, now)
			AdvanceEndpointQueues(endpoint, now)
			endpoint.Mutex.Unlock()
			assert.Equal(t, started.Add(tc.wantReset), endpoint.NextResetAt)
			for _, entries := range []map[string]*SlidingWindow{endpoint.UserCounts, endpoint.IpCounts, endpoint.RateLimitGroupCounts} {
				if tc.wantBuckets == 0 {
					assert.Empty(t, entries)
				} else if assert.NotNil(t, entries["identity"]) {
					assert.Equal(t, tc.wantBuckets, entries["identity"].Queue.Length())
					assert.Equal(t, 1, entries["identity"].Total)
				}
			}
		})
	}
}
