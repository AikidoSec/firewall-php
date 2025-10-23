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
	if Check(server, "", "GET", "/wp-config.php", nil) {
		t.Fatalf("expected no detection when IP is empty")
	}
}

func Test_NotAWebScanner(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	notScanner := []struct {
		method string
		route  string
	}{
		{"OPTIONS", "/"},
		{"GET", "/"},
		{"GET", "/login"},
		{"GET", "/dashboard"},
		{"GET", "/dashboard/2"},
		{"GET", "/settings"},
		{"GET", "/"},
		{"GET", "/dashboard"},
	}

	for i, req := range notScanner {
		if Check(server, ip, req.method, req.route, nil) {
			t.Fatalf("unexpected detection for non-scanner input at index %d: %s %s", i, req.method, req.route)
		}
	}
}

func Test_AWebScanner_ThresholdAndThrottle(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// First 5 scanner hits -> below threshold (6)
	paths := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
		"/.htaccess",
	}
	for i, p := range paths {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold at step %d", i)
		}
	}

	// 6th hit -> reaches threshold
	if !Check(server, ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("expected detection at 6th suspicious request")
	}

	// Immediately again -> throttled by minTimeBetweenEvents
	if Check(server, ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("unexpected detection due to throttle")
	}
}

func Test_AWebScanner_WithDelays(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious within the first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// "30s" delay in TS test -> still within same minute bucket for our window,
	// so we do not advance the minute here.

	// 5th (still below threshold)
	if Check(server, ip, "GET", "/.htaccess", nil) {
		t.Fatalf("unexpected detection at 5th request")
	}
	// 6th -> triggers
	if !Check(server, ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("expected detection at 6th request")
	}
	// throttled
	if Check(server, ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("unexpected detection due to throttle")
	}

	// "30 minutes later" in TS test -> still within 1h throttle.
	// Simulate by adjusting lastSentEvent back by 30m.
	server.AttackWaveMutex.Lock()
	server.AttackWaveLastSent[ip] = time.Now().Add(-30 * time.Minute)
	server.AttackWaveMutex.Unlock()

	// Still throttled (should short-circuit before counting)
	for i := 0; i < 3; i++ {
		if Check(server, ip, "GET", "/.env", nil) {
			t.Fatalf("unexpected detection while still within 1h throttle (iteration %d)", i)
		}
	}

	// "another 32 minutes" -> past 1h; also advance the window forward
	server.AttackWaveMutex.Lock()
	server.AttackWaveLastSent[ip] = time.Now().Add(-62 * time.Minute)
	server.AttackWaveMutex.Unlock()
	advanceN(server, 62) // clear old buckets

	// Now rebuild to threshold again
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before hitting 6 in new window at i=%d", i)
		}
	}
	if !Check(server, ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection after throttle elapsed and window re-accumulated")
	}
}

func Test_SlowScanner_TriggersInSecondInterval(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(server, 1)

	// Now build to threshold entirely within the second minute
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold in second interval at i=%d", i)
		}
	}
	if !Check(server, ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection at 6th request in second interval")
	}
}

func Test_SlowScanner_TriggersInThirdInterval(t *testing.T) {
	server := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(server, 1)

	// 4 more in second minute — still below threshold
	step2 := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
	}
	for i, p := range step2 {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step2[%d]", i)
		}
	}

	// Advance >1 minute again (TS tick another 62s)
	advanceN(server, 1)

	// Build to threshold in third interval
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if Check(server, ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold in third interval at i=%d", i)
		}
	}
	if !Check(server, ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection at 6th request in third interval")
	}
}
