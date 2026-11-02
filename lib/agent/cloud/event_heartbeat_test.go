package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatsAndClear(t *testing.T) {
	t.Run("reports operation counters without compressed timings and clears them", func(t *testing.T) {
		server := NewServerData()
		server.StatsData.MonitoredSinkTimings["PDO::query"] = MonitoredSinkTimings{
			Kind:                  "sql_op",
			AttacksDetected:       AttacksDetected{Total: 2, Blocked: 1},
			InterceptorThrewError: 1,
			WithoutContext:        3,
			Total:                 10,
		}

		stats := GetStatsAndClear(server)

		assert.Equal(t, map[string]MonitoredSinkStats{
			"PDO::query": {
				Kind:                  "sql_op",
				AttacksDetected:       AttacksDetected{Total: 2, Blocked: 1},
				InterceptorThrewError: 1,
				WithoutContext:        3,
				Total:                 10,
			},
		}, stats.Operations)
		payload, err := json.Marshal(Heartbeat{Stats: stats})
		require.NoError(t, err)
		assert.NotContains(t, string(payload), `"compressedTimings"`)
		assert.Empty(t, server.StatsData.MonitoredSinkTimings)
		assert.Empty(t, GetStatsAndClear(server).Operations)
	})

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
