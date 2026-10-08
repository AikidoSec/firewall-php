package context

import "testing"

func TestSetUserReplacesAndNormalizesUser(t *testing.T) {
	requestInstance := LoadForUnitTests(nil)
	defer UnloadForUnitTests()

	if !SetUser(requestInstance, "first\xff-id", "first\xfe-name") {
		t.Fatal("failed to set first user")
	}
	if id := GetUserId(requestInstance); id != "first-id" {
		t.Fatalf("expected normalized user ID, got %q", id)
	}
	if name := GetUserName(requestInstance); name != "first-name" {
		t.Fatalf("expected normalized user name, got %q", name)
	}

	if !SetUser(requestInstance, "second-id", "second-name") {
		t.Fatal("failed to replace user")
	}
	if id := GetUserId(requestInstance); id != "second-id" {
		t.Fatalf("expected replaced user ID, got %q", id)
	}
	if name := GetUserName(requestInstance); name != "second-name" {
		t.Fatalf("expected replaced user name, got %q", name)
	}
}

func TestTryTrackCustomEventLimitsEventsPerRequest(t *testing.T) {
	requestInstance := LoadForUnitTests(nil)
	defer UnloadForUnitTests()

	for i := 0; i < maxCustomEventsPerRequest; i++ {
		if !TryTrackCustomEvent(requestInstance) {
			t.Fatalf("expected event %d to be accepted", i+1)
		}
	}
	if TryTrackCustomEvent(requestInstance) {
		t.Fatal("expected event over the per-request limit to be dropped")
	}

	ctx := GetContext(requestInstance)
	if ctx.customEventsTracked != maxCustomEventsPerRequest {
		t.Fatalf("expected %d tracked events, got %d", maxCustomEventsPerRequest, ctx.customEventsTracked)
	}
	if !ctx.customEventLimitWarningLogged {
		t.Fatal("expected the event limit warning to be marked as logged")
	}

	Clear(requestInstance)
	if !TryTrackCustomEvent(requestInstance) {
		t.Fatal("expected a new request to have a fresh event allowance")
	}
}
