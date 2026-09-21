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
