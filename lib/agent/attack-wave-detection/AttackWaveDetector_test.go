package attackwavedetection

import (
	"testing"
	"time"
)

// Helpers
// add this helper method anywhere on the type (e.g., near advanceByElapsed)
// test helper: advance all queues by one minute bucket
func (a *AttackWaveDetector) advanceAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, q := range a.ipQueues {
		q.pushNewMinute()
	}
	// keep lastTick roughly in step, so subsequent elapsed advances behave
	if a.lastTick.IsZero() {
		a.lastTick = time.Now()
	} else {
		a.lastTick = a.lastTick.Add(a.tickDur)
	}
}

func newDetectorForTSDefaults() *AttackWaveDetector {
	// TS tests used:
	// threshold = 6, timeframe 60s, minBetween = 1h
	// Our window is per-minute buckets with size=1 => equivalent for these flows.
	return New(Options{
		AttackWaveThreshold:  6,
		WindowSizeInMinutes:  1,
		MinTimeBetweenEvents: time.Hour,
		MaxEntries:           10000,
		// We won't Start() the ticker; tests call advanceAll() directly.
	})
}

func advanceN(d *AttackWaveDetector, n int) {
	for i := 0; i < n; i++ {
		d.advanceAll()
	}
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func Test_NoIPAddress(t *testing.T) {
	d := newDetectorForTSDefaults()
	if d.Check("", "GET", "/wp-config.php", nil) {
		t.Fatalf("expected no detection when IP is empty")
	}
}

func Test_NotAWebScanner(t *testing.T) {
	d := newDetectorForTSDefaults()
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
		if d.Check(ip, req.method, req.route, nil) {
			t.Fatalf("unexpected detection for non-scanner input at index %d: %s %s", i, req.method, req.route)
		}
	}
}

func Test_AWebScanner_ThresholdAndThrottle(t *testing.T) {
	d := newDetectorForTSDefaults()
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
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold at step %d", i)
		}
	}

	// 6th hit -> reaches threshold
	if !d.Check(ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("expected detection at 6th suspicious request")
	}

	// Immediately again -> throttled by minTimeBetweenEvents
	if d.Check(ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("unexpected detection due to throttle")
	}
}

func Test_AWebScanner_WithDelays(t *testing.T) {
	d := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious within the first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// "30s" delay in TS test -> still within same minute bucket for our window,
	// so we do not advance the minute here.

	// 5th (still below threshold)
	if d.Check(ip, "GET", "/.htaccess", nil) {
		t.Fatalf("unexpected detection at 5th request")
	}
	// 6th -> triggers
	if !d.Check(ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("expected detection at 6th request")
	}
	// throttled
	if d.Check(ip, "GET", "/.htpasswd", nil) {
		t.Fatalf("unexpected detection due to throttle")
	}

	// "30 minutes later" in TS test -> still within 1h throttle.
	// Simulate by adjusting lastSentEvent back by 30m.
	d.mu.Lock()
	d.lastSentEvent[ip] = time.Now().Add(-30 * time.Minute)
	d.mu.Unlock()

	// Still throttled (should short-circuit before counting)
	for i := 0; i < 3; i++ {
		if d.Check(ip, "GET", "/.env", nil) {
			t.Fatalf("unexpected detection while still within 1h throttle (iteration %d)", i)
		}
	}

	// "another 32 minutes" -> past 1h; also advance the window forward
	d.mu.Lock()
	d.lastSentEvent[ip] = time.Now().Add(-62 * time.Minute)
	d.mu.Unlock()
	advanceN(d, 62) // clear old buckets

	// Now rebuild to threshold again
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before hitting 6 in new window at i=%d", i)
		}
	}
	if !d.Check(ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection after throttle elapsed and window re-accumulated")
	}
}

func Test_SlowScanner_TriggersInSecondInterval(t *testing.T) {
	d := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(d, 1)

	// Now build to threshold entirely within the second minute
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold in second interval at i=%d", i)
		}
	}
	if !d.Check(ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection at 6th request in second interval")
	}
}

func Test_SlowScanner_TriggersInThirdInterval(t *testing.T) {
	d := newDetectorForTSDefaults()
	ip := "::1"

	// 4 suspicious in first minute
	step1 := []string{
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range step1 {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step1[%d]", i)
		}
	}

	// Advance >1 minute (TS tick 62s)
	advanceN(d, 1)

	// 4 more in second minute — still below threshold
	step2 := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
	}
	for i, p := range step2 {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection at step2[%d]", i)
		}
	}

	// Advance >1 minute again (TS tick another 62s)
	advanceN(d, 1)

	// Build to threshold in third interval
	reqs := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/.git/config",
		"/.env",
	}
	for i, p := range reqs {
		if d.Check(ip, "GET", p, nil) {
			t.Fatalf("unexpected detection before threshold in third interval at i=%d", i)
		}
	}
	if !d.Check(ip, "GET", "/.htaccess", nil) {
		t.Fatalf("expected detection at 6th request in third interval")
	}
}
