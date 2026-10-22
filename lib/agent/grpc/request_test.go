package grpc

import (
	"sync"
	"testing"
	"testing/synctest"
	"time"

	. "main/aikido_types"
	"main/ipc/protos"
	"main/log"
	"main/rate_limiting"
	"main/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitingCatchesUpAfterDelayedTicks(t *testing.T) {
	for _, tc := range []struct {
		name, user, group string
		tickFirst         bool
	}{
		{"IP/request first", "", "", false},
		{"IP/ticker first", "", "", true},
		{"user/request first", "alice", "", false},
		{"user/ticker first", "alice", "", true},
		{"group/request first", "alice", "team", false},
		{"group/ticker first", "alice", "team", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				endpoint := &RateLimitingValue{
					Config:               RateLimitingConfig{MaxRequests: 1, WindowSizeInMinutes: 3},
					UserCounts:           make(map[string]*SlidingWindow),
					IpCounts:             make(map[string]*SlidingWindow),
					RateLimitGroupCounts: make(map[string]*SlidingWindow),
					NextResetAt:          time.Now().Add(time.Minute),
				}
				server := &ServerData{
					Logger: log.CreateLogger("test", "ERROR", false),
					RateLimitingMap: map[RateLimitingKey]*RateLimitingValue{
						{Method: "GET", Route: "/"}: endpoint,
					},
				}
				status := func() *protos.RateLimitingStatus {
					return getRateLimitingStatus(server, "GET", "/", "/", tc.user, "203.0.113.1", tc.group)
				}
				require.False(t, status().Block)
				require.Equal(t, int64(180), status().RetryAfter)
				time.Sleep(55 * time.Second)
				rate_limiting.AdvanceRateLimitingQueues(server)
				require.Equal(t, int64(125), status().RetryAfter, "cleanup must wait for the endpoint's deadline")
				time.Sleep(135 * time.Second)
				if tc.tickFirst {
					rate_limiting.AdvanceRateLimitingQueues(server)
				}
				require.False(t, status().Block, "expired requests must not keep blocking")
				rate_limiting.AdvanceRateLimitingQueues(server)
				blocked := status()
				require.True(t, blocked.Block, "the delayed callback must preserve the new request")
				require.Equal(t, int64(170), blocked.RetryAfter)
				time.Sleep(time.Duration(blocked.RetryAfter) * time.Second)
				require.False(t, status().Block, "retry must succeed without waiting for a ticker callback")
			})
		})
	}
}

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
		assert.False(t, updateAttackWaveCountsAndDetect(server, &server.AttackWave, &server.AttackWaveMutex, true, ip, "", "", "", "", "", ""))
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
		assert.True(t, updateAttackWaveCountsAndDetect(server, &server.AttackWave, &server.AttackWaveMutex, true, ip, "", "", "", "", "", ""))

		// Verify LastSent map was populated
		assert.True(t, server.AttackWave.LastSent[ip] > 0, "LastSent should be set after event is sent")

		// Verify the attack wave stats counter was incremented, so it gets reported in the heartbeat stats
		assert.Equal(t, 1, server.StatsData.AttackWaves)
	})
}

func TestGetRuntimeAnomalySampleReason(t *testing.T) {
	assert.Equal(t, "", getRuntimeAnomalySampleReason("", false))
	assert.Equal(t, "memory_exhausted", getRuntimeAnomalySampleReason("memory_exhausted", false))
	assert.Equal(t, "execution_timeout", getRuntimeAnomalySampleReason("execution_timeout", true))
	assert.Equal(t, "fatal_error", getRuntimeAnomalySampleReason("fatal_error", false))
	assert.Equal(t, "session_upload_progress", getRuntimeAnomalySampleReason("", true))
}

func newRuntimeAnomalyTestServer(threshold int) *ServerData {
	return &ServerData{
		Logger: log.CreateLogger("test", "ERROR", false),
		AttackWave: AttackWaveState{
			Threshold:       threshold,
			WindowSize:      1,
			MinBetween:      60000,
			IpQueues:        make(map[string]*SlidingWindow),
			LastSent:        make(map[string]int64),
			MaxSamplesPerIP: 15,
		},
		RuntimeAnomalyWave: AttackWaveState{
			Threshold:       threshold,
			WindowSize:      1,
			MinBetween:      60000,
			IpQueues:        make(map[string]*SlidingWindow),
			LastSent:        make(map[string]int64),
			MaxSamplesPerIP: 15,
		},
	}
}

func TestRuntimeAnomalyWave(t *testing.T) {
	detect := func(server *ServerData, ip, url, reason string) bool {
		return updateAttackWaveCountsAndDetect(server, &server.RuntimeAnomalyWave, &server.RuntimeAnomalyWaveMutex, reason != "", ip, "", "", "POST", url, reason, runtimeAnomaliesEventReason)
	}

	t.Run("ignores requests without a runtime anomaly", func(t *testing.T) {
		server := newRuntimeAnomalyTestServer(1)

		assert.False(t, detect(server, "1.2.3.4", "/upload", ""))
		assert.Empty(t, server.RuntimeAnomalyWave.IpQueues)
	})

	t.Run("detects at the threshold and records the reason of each sample", func(t *testing.T) {
		server := newRuntimeAnomalyTestServer(3)
		ip := "1.2.3.4"

		assert.False(t, detect(server, ip, "/upload", "memory_exhausted"))
		assert.False(t, detect(server, ip, "/upload", "session_upload_progress"))
		assert.True(t, detect(server, ip, "/import", "execution_timeout"))

		samples := server.RuntimeAnomalyWave.IpQueues[ip].Samples
		assert.Equal(t, []SuspiciousRequest{
			{Method: "POST", Url: "/upload", Reason: "memory_exhausted"},
			{Method: "POST", Url: "/upload", Reason: "session_upload_progress"},
			{Method: "POST", Url: "/import", Reason: "execution_timeout"},
		}, samples)
		assert.True(t, server.RuntimeAnomalyWave.LastSent[ip] > 0)
		assert.Equal(t, 1, server.StatsData.AttackWaves)
	})

	t.Run("does not affect web scanner attack wave state", func(t *testing.T) {
		server := newRuntimeAnomalyTestServer(1)
		ip := "1.2.3.4"

		assert.True(t, detect(server, ip, "/upload", "fatal_error"))
		assert.Empty(t, server.AttackWave.IpQueues)
		assert.Empty(t, server.AttackWave.LastSent)
	})

	t.Run("is throttled per IP after an event was sent", func(t *testing.T) {
		server := newRuntimeAnomalyTestServer(1)
		ip := "1.2.3.4"

		assert.True(t, detect(server, ip, "/upload", "memory_exhausted"))
		assert.False(t, detect(server, ip, "/upload", "memory_exhausted"))
		assert.True(t, detect(server, "5.6.7.8", "/upload", "memory_exhausted"))
		assert.Equal(t, 2, server.StatsData.AttackWaves)
	})
}
