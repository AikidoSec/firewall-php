package grpc

import (
	"sync"
	"testing"

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
