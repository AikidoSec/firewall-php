package rate_limiting

import (
	. "main/aikido_types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdvanceTracksNextReset(t *testing.T) {
	counts := func() map[string]*SlidingWindow {
		window := NewSlidingWindow()
		window.Increment()
		return map[string]*SlidingWindow{"identity": window}
	}
	endpoint := &RateLimitingValue{
		Config:     RateLimitingConfig{MaxRequests: 1, WindowSizeInMinutes: 2},
		UserCounts: counts(), IpCounts: counts(), RateLimitGroupCounts: counts(),
	}
	server := &ServerData{RateLimitingMap: map[RateLimitingKey]*RateLimitingValue{
		{Method: "GET", Route: "/"}: endpoint,
	}}
	tick := time.Unix(1000, 0)
	advanceRateLimitingQueues(server, tick, tick.Add(250*time.Millisecond))
	assert.Equal(t, tick.Add(time.Minute), server.RateLimitingNextResetAt)
	for _, entries := range []map[string]*SlidingWindow{endpoint.UserCounts, endpoint.IpCounts, endpoint.RateLimitGroupCounts} {
		assert.Equal(t, int64(5), entries["identity"].RetryAfter(2, 1, server.RateLimitingNextResetAt, tick.Add(55*time.Second)))
	}
	// A delayed callback still advances one bucket, matching the existing limiter.
	advanceRateLimitingQueues(server, tick.Add(time.Minute), tick.Add(150*time.Second))
	assert.Equal(t, tick.Add(3*time.Minute), server.RateLimitingNextResetAt)
	assert.Empty(t, endpoint.UserCounts)
	assert.Empty(t, endpoint.IpCounts)
	assert.Empty(t, endpoint.RateLimitGroupCounts)
}
