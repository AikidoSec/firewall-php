package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduleCustomEventDropsEventsWhileCloudIsSlow(t *testing.T) {
	var received atomic.Int32
	release := make(chan struct{})
	cloudServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		received.Add(1)
		<-release
		writer.WriteHeader(http.StatusOK)
	}))
	defer cloudServer.Close()
	var releaseOnce sync.Once
	releaseCloud := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseCloud()

	server := &ServerData{
		Logger: log.CreateLogger("test", "ERROR", false),
		AikidoConfig: AikidoConfigData{
			Token:    "AIK_TEST",
			Endpoint: cloudServer.URL,
		},
	}

	scheduled := make(chan struct{})
	go func() {
		for i := 0; i < constants.MaxConcurrentCustomEventRequests+1; i++ {
			ScheduleCustomEvent(server, CustomEvent{Type: "custom", Name: "user.login_failed"})
		}
		close(scheduled)
	}()
	select {
	case <-scheduled:
	case <-time.After(time.Second):
		t.Fatal("scheduling custom events must not wait for the cloud")
	}

	if !waitFor(5*time.Second, func() bool { return received.Load() == constants.MaxConcurrentCustomEventRequests }) {
		t.Fatalf("expected %d requests to reach the cloud, got %d", constants.MaxConcurrentCustomEventRequests, received.Load())
	}
	releaseCloud()
	if !waitFor(5*time.Second, func() bool { return len(customEventSlots) == 0 }) {
		t.Fatal("expected all custom event slots to be released")
	}
	if received.Load() != constants.MaxConcurrentCustomEventRequests {
		t.Fatalf("expected the event over the limit to be dropped, got %d requests", received.Load())
	}

	ScheduleCustomEvent(server, CustomEvent{Type: "custom", Name: "user.login_failed"})
	if !waitFor(5*time.Second, func() bool { return received.Load() == constants.MaxConcurrentCustomEventRequests+1 }) {
		t.Fatal("expected a released slot to accept a new custom event")
	}
	if !waitFor(5*time.Second, func() bool { return len(customEventSlots) == 0 }) {
		t.Fatal("expected all custom event slots to be released")
	}
}
