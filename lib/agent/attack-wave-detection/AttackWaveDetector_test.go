package attackwavedetection

import (
	. "main/aikido_types"
	"main/log"
	"testing"
	"time"
)

// Helpers
// test helper: advance all queues by one minute bucket
func advanceAll(server *ServerData) {
	server.AttackWaveMutex.Lock()
	defer server.AttackWaveMutex.Unlock()
	for _, q := range server.AttackWaveIpQueues {
		pushNewMinute(q)
	}
	// keep lastTick roughly in step, so subsequent elapsed advances behave
	tickDur := 1 * time.Minute
	if server.AttackWaveLastTick.IsZero() {
		server.AttackWaveLastTick = time.Now()
	} else {
		server.AttackWaveLastTick = server.AttackWaveLastTick.Add(tickDur)
	}
}

func newDetectorForTSDefaults() *ServerData {
	// TS tests used:
	// threshold = 6, timeframe 60s, minBetween = 1h
	// Our window is per-minute buckets with size=1 => equivalent for these flows.

	// Initialize MainLogger if not already initialized
	if log.MainLogger == nil {
		log.MainLogger = log.CreateLogger("test", "INFO", false)
	}

	server := NewServerData()
	server.Logger = log.MainLogger
	server.AttackWaveThreshold = 6
	server.AttackWaveWindowSize = 1
	server.AttackWaveMinBetween = time.Hour
	server.AttackWaveMaxEntries = 10000
	return server
}

func advanceN(server *ServerData, n int) {
	for i := 0; i < n; i++ {
		advanceAll(server)
	}
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func Test_NoIPAddress(t *testing.T) {
	server := newDetectorForTSDefaults()
	if IncrementAndDetect(server, "") {
		t.Fatalf("expected no detection when IP is empty")
	}
}

func Test_AWebScanner_ThresholdAndThrottle(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	for i := range 5 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection before threshold at step %d", i)
		}
	}

	// 6th hit -> reaches threshold
	if !IncrementAndDetect(server, ip) {
		t.Fatalf("expected detection at 6th suspicious request")
	}

	// Immediately again -> throttled by minTimeBetweenEvents
	if IncrementAndDetect(server, ip) {
		t.Fatalf("unexpected detection due to throttle")
	}
}

func Test_AWebScanner_WithDelays(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious within the first minute
	for i := range 4 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// "30s" delay in TS test -> still within same minute bucket for our window,
	// so we do not advance the minute here.

	// 5th (still below threshold)
	if IncrementAndDetect(server, ip) {
		t.Fatalf("unexpected detection at 5th request")
	}
	// 6th -> triggers
	if !IncrementAndDetect(server, ip) {
		t.Fatalf("expected detection at 6th request")
	}
	// throttled
	if IncrementAndDetect(server, ip) {
		t.Fatalf("unexpected detection due to throttle")
	}

	// "30 minutes later" in TS test -> still within 1h throttle.
	// Simulate by adjusting lastSentEvent back by 30m.
	server.AttackWaveMutex.Lock()
	server.AttackWaveLastSent[ip] = time.Now().Add(-30 * time.Minute)
	server.AttackWaveMutex.Unlock()

	// Still throttled (should short-circuit before counting)
	for i := 0; i < 3; i++ {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection while still within 1h throttle (iteration %d)", i)
		}
	}

	// "another 32 minutes" -> past 1h; also advance the window forward
	server.AttackWaveMutex.Lock()
	server.AttackWaveLastSent[ip] = time.Now().Add(-62 * time.Minute)
	server.AttackWaveMutex.Unlock()
	advanceN(server, 62) // clear old buckets

	// Now rebuild to threshold again
	for i := range 5 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection before hitting 6 in new window at i=%d", i)
		}
	}
	if !IncrementAndDetect(server, ip) {
		t.Fatalf("expected detection after throttle elapsed and window re-accumulated")
	}
}

func Test_SlowScanner_TriggersInSecondInterval(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	for i := range 4 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(server, 1)

	// Now build to threshold entirely within the second minute
	for i := range 5 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection before threshold in second interval at i=%d", i)
		}
	}
	if !IncrementAndDetect(server, ip) {
		t.Fatalf("expected detection at 6th request in second interval")
	}
}

func Test_SlowScanner_TriggersInThirdInterval(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	for i := range 4 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(server, 1)

	// 4 more in second minute — still below threshold
	for i := range 4 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection at step2[%d]", i)
		}
	}

	// Advance >1 minute again (TS tick another 62s)
	advanceN(server, 1)

	// Build to threshold in third interval
	for i := range 5 {
		if IncrementAndDetect(server, ip) {
			t.Fatalf("unexpected detection before threshold in third interval at i=%d", i)
		}
	}
	if !IncrementAndDetect(server, ip) {
		t.Fatalf("expected detection at 6th request in third interval")
	}
}
