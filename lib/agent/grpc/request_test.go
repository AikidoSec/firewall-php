package grpc

import (
	"sync"
	"testing"
	"time"

	. "main/aikido_types"
	"main/ipc/protos"
	"main/log"
	"main/utils"

	"github.com/stretchr/testify/assert"
)

func TestStoreTotalStats(t *testing.T) {
	t.Run("increments Requests on every call, and RequestsRateLimited only when rate limited", func(t *testing.T) {
		server := &ServerData{}

		storeTotalStats(server, false)
		assert.Equal(t, 1, server.StatsData.Requests)
		assert.Equal(t, 0, server.StatsData.RequestsRateLimited)

		storeTotalStats(server, true)
		assert.Equal(t, 2, server.StatsData.Requests)
		assert.Equal(t, 1, server.StatsData.RequestsRateLimited)

		storeTotalStats(server, true)
		assert.Equal(t, 3, server.StatsData.Requests)
		assert.Equal(t, 2, server.StatsData.RequestsRateLimited)
	})
}

func TestStoreAttackStats(t *testing.T) {
	t.Run("increments Attacks on every call, and AttacksBlocked only when blocked", func(t *testing.T) {
		server := &ServerData{}

		storeAttackStats(server, &protos.AttackDetected{Attack: &protos.Attack{Blocked: false}})
		assert.Equal(t, 1, server.StatsData.Attacks)
		assert.Equal(t, 0, server.StatsData.AttacksBlocked)

		storeAttackStats(server, &protos.AttackDetected{Attack: &protos.Attack{Blocked: true}})
		assert.Equal(t, 2, server.StatsData.Attacks)
		assert.Equal(t, 1, server.StatsData.AttacksBlocked)

		storeAttackStats(server, &protos.AttackDetected{Attack: &protos.Attack{Blocked: true}})
		assert.Equal(t, 3, server.StatsData.Attacks)
		assert.Equal(t, 2, server.StatsData.AttacksBlocked)
	})
}

func TestStoreAttackWaveStats(t *testing.T) {
	t.Run("increments AttackWaves on every call", func(t *testing.T) {
		server := &ServerData{}

		storeAttackWaveStats(server)
		assert.Equal(t, 1, server.StatsData.AttackWaves)

		storeAttackWaveStats(server)
		storeAttackWaveStats(server)
		assert.Equal(t, 3, server.StatsData.AttackWaves)

		assert.Equal(t, 0, server.StatsData.AttackWavesBlocked)
	})

	t.Run("is safe for concurrent use", func(t *testing.T) {
		server := &ServerData{}

		const goroutines = 50
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				storeAttackWaveStats(server)
			}()
		}
		wg.Wait()

		assert.Equal(t, goroutines, server.StatsData.AttackWaves)
	})
}

func TestAttackWaveThrottling(t *testing.T) {
	t.Run("returns false when event for IP was recently sent (within MinBetween window)", func(t *testing.T) {
		server := &ServerData{
			AttackWave: AttackWaveState{
				Threshold:  10,
				WindowSize: 20,
				MinBetween: 60000, // 60 seconds in milliseconds
				IpQueues:   make(map[string]*SlidingWindow),
				LastSent:   make(map[string]int64),
			},
			AttackWaveMutex: sync.Mutex{},
		}

		ip := "192.168.1.1"
		now := utils.GetTime()

		// Manually set LastSent time to recent past (30 seconds ago)
		server.AttackWave.LastSent[ip] = now - 30000

		// Create a sliding window for this IP with counts above threshold
		sw := NewSlidingWindow()
		for i := 0; i < 10; i++ {
			sw.Increment()
		}
		server.AttackWave.IpQueues[ip] = sw

		// Should return false (throttled) because last event was only 30 seconds ago (< 60s MinBetween)
		assert.False(t, updateAttackWaveCountsAndDetect(server, true, ip, "", "", "", ""))
	})

	t.Run("returns true and populates LastSent map when IP reaches threshold for first time", func(t *testing.T) {
		server := &ServerData{
			Logger: log.CreateLogger("test", "ERROR", false),
			AttackWave: AttackWaveState{
				Threshold:  10,
				WindowSize: 20,
				MinBetween: 60000, // 60 seconds in milliseconds
				IpQueues:   make(map[string]*SlidingWindow),
				LastSent:   make(map[string]int64),
			},
			AttackWaveMutex: sync.Mutex{},
		}

		ip := "192.168.1.1"

		// Create a sliding window for this IP with counts above threshold
		sw := NewSlidingWindow()
		for i := 0; i < 9; i++ {
			sw.Increment()
		}
		server.AttackWave.IpQueues[ip] = sw

		// Verify IP doesn't exist in LastSent map initially
		_, exists := server.AttackWave.LastSent[ip]
		assert.False(t, exists, "IP should not be in LastSent map before threshold is reached")

		// Should return true (event sent) because this is the first time reaching threshold
		assert.True(t, updateAttackWaveCountsAndDetect(server, true, ip, "", "", "", ""))

		// Verify LastSent map was populated
		assert.True(t, server.AttackWave.LastSent[ip] > 0, "LastSent should be set after event is sent")

		// Verify the attack wave stats counter was incremented, so it gets reported in the heartbeat stats
		assert.Equal(t, 1, server.StatsData.AttackWaves)
	})
}

func TestComputeRetryAfterSeconds(t *testing.T) {
	t.Run("returns full window size when sliding window is nil", func(t *testing.T) {
		result := computeRetryAfterSeconds(nil, 5)
		assert.Equal(t, int32(300), result)
	})

	t.Run("returns full window size for a freshly created window", func(t *testing.T) {
		sw := NewSlidingWindow()
		result := computeRetryAfterSeconds(sw, 5)
		assert.True(t, result >= 299 && result <= 300, "expected ~300, got %d", result)
	})

	t.Run("decreases over time", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.CreatedAt = time.Now().Add(-60 * time.Second)
		result := computeRetryAfterSeconds(sw, 5)
		assert.True(t, result >= 239 && result <= 241, "expected ~240, got %d", result)
	})

	t.Run("clamps to 1 when window has expired", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.CreatedAt = time.Now().Add(-600 * time.Second)
		result := computeRetryAfterSeconds(sw, 5)
		assert.Equal(t, int32(1), result)
	})

	t.Run("stays accurate after CreatedAt is advanced by eviction", func(t *testing.T) {
		sw := NewSlidingWindow()
		sw.CreatedAt = time.Now().Add(-30 * time.Second)

		// Simulate one eviction advancing CreatedAt by 1 minute
		sw.CreatedAt = sw.CreatedAt.Add(time.Minute)

		result := computeRetryAfterSeconds(sw, 2) // 2 min window = 120s
		// CreatedAt is now ~30 seconds in the future, so retryAfter ≈ 120 + 30 = 150
		assert.True(t, result >= 149 && result <= 151, "expected ~150, got %d", result)
	})
}
