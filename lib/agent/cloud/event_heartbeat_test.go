package cloud

import (
	. "main/aikido_types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatsAndClear(t *testing.T) {
	t.Run("reports accumulated counters and clears them for the next interval", func(t *testing.T) {
		server := NewServerData()
		server.StatsData.Requests = 30
		server.StatsData.RequestsAborted = 2
		server.StatsData.RequestsRateLimited = 7
		server.StatsData.Attacks = 5
		server.StatsData.AttacksBlocked = 3
		server.StatsData.AttackWaves = 2
		server.StatsData.AttackWavesBlocked = 1

		stats := GetStatsAndClear(server)

		// The returned snapshot reflects everything that was accumulated (incrementing side)
		assert.Equal(t, 30, stats.Requests.Total)
		assert.Equal(t, 2, stats.Requests.Aborted)
		assert.Equal(t, 7, stats.Requests.RateLimited)
		assert.Equal(t, 5, stats.Requests.AttacksDetected.Total)
		assert.Equal(t, 3, stats.Requests.AttacksDetected.Blocked)
		assert.Equal(t, 2, stats.Requests.AttackWaves.Total)
		assert.Equal(t, 1, stats.Requests.AttackWaves.Blocked)

		// The underlying counters are cleared (decremented back to zero) so the next heartbeat starts fresh
		assert.Equal(t, 0, server.StatsData.Requests)
		assert.Equal(t, 0, server.StatsData.RequestsAborted)
		assert.Equal(t, 0, server.StatsData.RequestsRateLimited)
		assert.Equal(t, 0, server.StatsData.Attacks)
		assert.Equal(t, 0, server.StatsData.AttacksBlocked)
		assert.Equal(t, 0, server.StatsData.AttackWaves)
		assert.Equal(t, 0, server.StatsData.AttackWavesBlocked)
	})

	t.Run("a second call after no new activity reports zeroes", func(t *testing.T) {
		server := NewServerData()
		server.StatsData.Attacks = 4
		server.StatsData.AttackWaves = 4

		_ = GetStatsAndClear(server)
		stats := GetStatsAndClear(server)

		assert.Equal(t, 0, stats.Requests.AttacksDetected.Total)
		assert.Equal(t, 0, stats.Requests.AttackWaves.Total)
	})

	t.Run("counters incremented after a clear are picked up by the next snapshot", func(t *testing.T) {
		server := NewServerData()
		server.StatsData.AttackWaves = 1

		first := GetStatsAndClear(server)
		assert.Equal(t, 1, first.Requests.AttackWaves.Total)

		server.StatsData.AttackWaves += 2

		second := GetStatsAndClear(server)
		assert.Equal(t, 2, second.Requests.AttackWaves.Total)
	})
}
