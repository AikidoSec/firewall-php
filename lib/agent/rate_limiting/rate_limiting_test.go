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
	server.RateLimitingStartedAt = tick
	nextReset := NextResetAt(server, tick.Add(250*time.Millisecond))
	assert.Equal(t, tick.Add(time.Minute), nextReset)
	advanceRateLimitingQueues(server, nextReset)
	assert.Equal(t, nextReset, endpoint.NextResetAt)
	for _, entries := range []map[string]*SlidingWindow{endpoint.UserCounts, endpoint.IpCounts, endpoint.RateLimitGroupCounts} {
		assert.Equal(t, int64(5), entries["identity"].RetryAfter(2, 1, endpoint.NextResetAt, tick.Add(55*time.Second)))
	}
	// An endpoint already initialized for this minute must not rotate again.
	advanceRateLimitingQueues(server, nextReset)
	assert.Equal(t, 1, endpoint.UserCounts["identity"].Total)

	// A delayed callback still advances one bucket, matching the existing limiter.
	nextReset = NextResetAt(server, tick.Add(150*time.Second))
	assert.Equal(t, tick.Add(3*time.Minute), nextReset)
	advanceRateLimitingQueues(server, nextReset)
	assert.Equal(t, nextReset, endpoint.NextResetAt)
	assert.Empty(t, endpoint.UserCounts)
	assert.Empty(t, endpoint.IpCounts)
	assert.Empty(t, endpoint.RateLimitGroupCounts)
}
