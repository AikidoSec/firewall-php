package cloud

import (
	"context"
	"fmt"
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/* Mock of the Aikido cloud, serving the config stream and the config endpoints used after an event */
type mockCloud struct {
	httpServer *httptest.Server

	// configUpdatedAt is the timestamp returned by the config endpoint
	configUpdatedAt atomic.Int64

	// pushEvent sends a config-updated event with the given timestamp to a connected stream
	pushEvent chan int64

	// streamStatus is the status code the stream endpoint replies with
	streamStatus atomic.Int32

	// closeStreamImmediately makes the stream endpoint reply 200 and hang up right away
	closeStreamImmediately atomic.Bool

	streamConnections atomic.Int32
	streamDisconnects atomic.Int32
	configRequests    atomic.Int32
	streamHeaders     atomic.Value
}

func newMockCloud() *mockCloud {
	cloud := &mockCloud{pushEvent: make(chan int64)}
	cloud.streamStatus.Store(http.StatusOK)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/runtime/stream", func(w http.ResponseWriter, r *http.Request) {
		cloud.streamConnections.Add(1)
		cloud.streamHeaders.Store(r.Header.Clone())

		if status := int(cloud.streamStatus.Load()); status != http.StatusOK {
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		if cloud.closeStreamImmediately.Load() {
			return
		}

		flusher := w.(http.Flusher)
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				cloud.streamDisconnects.Add(1)
				return
			case configUpdatedAt := <-cloud.pushEvent:
				fmt.Fprintf(w, "event: config-updated\ndata: {\"configUpdatedAt\":%d}\n\n", configUpdatedAt)
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("/api/runtime/config", func(w http.ResponseWriter, r *http.Request) {
		cloud.configRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"serviceId":1,"configUpdatedAt":%d,"heartbeatIntervalInMS":600000,
			"endpoints":[],"blockedUserIds":[],"allowedIPAddresses":[],"receivedAnyStats":true,"block":true}`,
			cloud.configUpdatedAt.Load())
	})

	mux.HandleFunc("/api/runtime/firewall/lists", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success":true,"serviceId":1}`)
	})

	cloud.httpServer = httptest.NewServer(mux)
	return cloud
}

/* Pushes a config-updated event, first making the config endpoint report the new timestamp */
func (cloud *mockCloud) publishConfig(t *testing.T, configUpdatedAt int64) {
	cloud.configUpdatedAt.Store(configUpdatedAt)

	select {
	case cloud.pushEvent <- configUpdatedAt:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the config stream to pick up the event")
	}
}

func newTestServerData(cloud *mockCloud, token string, sse bool) *ServerData {
	server := NewServerData()
	server.Logger = log.CreateLogger("test", "ERROR", false)
	server.AikidoConfig.Token = token
	server.AikidoConfig.Endpoint = cloud.httpServer.URL
	server.AikidoConfig.ConfigEndpoint = cloud.httpServer.URL
	server.AikidoConfig.Sse = sse
	return server
}

func getConfigUpdatedAt(server *ServerData) int64 {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	return server.CloudConfig.ConfigUpdatedAt
}

func waitFor(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return condition()
}

func TestConfigStreamRoutine(t *testing.T) {
	t.Run("it applies a config pushed over the stream", func(t *testing.T) {
		cloud := newMockCloud()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		StartConfigStreamRoutine(server)

		// The stream must be stopped before the mock cloud is closed, as closing waits for open requests
		stopRoutine := sync.OnceFunc(func() { StopConfigStreamRoutine(server) })
		defer cloud.httpServer.Close()
		defer stopRoutine()

		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamConnections.Load() == 1
		}), "expected the agent to connect to the config stream")

		streamHeaders := cloud.streamHeaders.Load().(http.Header)
		assert.Equal(t, "AIK_RUNTIME_TEST_TOKEN", streamHeaders.Get("Authorization"))
		assert.Equal(t, "text/event-stream", streamHeaders.Get("Accept"))
		assert.Equal(t, "no-cache", streamHeaders.Get("Cache-Control"))
		assert.Equal(t, "identity", streamHeaders.Get("Accept-Encoding"))
		assert.Equal(t, "php", streamHeaders.Get("X-Agent-Platform"))
		assert.Equal(t, constants.Version, streamHeaders.Get("X-Agent-Version"))

		cloud.publishConfig(t, 1000)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return getConfigUpdatedAt(server) == 1000
		}), "expected the pushed config to be applied")
		assert.Equal(t, int32(1), cloud.configRequests.Load())

		// A second event for a config we already have must not trigger another fetch
		cloud.publishConfig(t, 1000)
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, int32(1), cloud.configRequests.Load())

		cloud.publishConfig(t, 2000)
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, int64(1000), getConfigUpdatedAt(server))
		assert.Equal(t, int32(1), cloud.configRequests.Load())

		server.ConfigStreamRefreshMutex.Lock()
		server.ConfigStreamLastRefreshStart = time.Now().Add(-9 * time.Second)
		server.ConfigStreamRefreshMutex.Unlock()

		cloud.publishConfig(t, 3000)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return getConfigUpdatedAt(server) == 3000
		}), "expected the throttled config stream to refresh again")
		assert.Equal(t, int32(2), cloud.configRequests.Load())

		// The connection must be dropped when the routine is stopped
		stopRoutine()
		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamDisconnects.Load() == 1
		}), "expected the config stream connection to be closed on stop")
	})

	t.Run("it does not connect when the feature is disabled", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", false)

		StartConfigStreamRoutine(server)
		defer StopConfigStreamRoutine(server)

		time.Sleep(300 * time.Millisecond)
		assert.Equal(t, int32(0), cloud.streamConnections.Load())
	})

	t.Run("it connects when the cloud enables realtime updates", func(t *testing.T) {
		cloud := newMockCloud()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", false)
		server.CloudConfig.EnabledFeatures = []string{constants.RealtimeUpdatesFeature}

		StartConfigStreamRoutine(server)

		stopRoutine := sync.OnceFunc(func() { StopConfigStreamRoutine(server) })
		defer cloud.httpServer.Close()
		defer stopRoutine()

		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamConnections.Load() == 1
		}), "expected the agent to connect when the cloud enables the feature")
	})

	t.Run("it does not connect for a cloud feature it does not know", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", false)
		server.CloudConfig.EnabledFeatures = []string{"some_other_feature"}

		StartConfigStreamRoutine(server)
		defer StopConfigStreamRoutine(server)

		time.Sleep(300 * time.Millisecond)
		assert.Equal(t, int32(0), cloud.streamConnections.Load())
	})

	t.Run("it does not connect without a token", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		server := newTestServerData(cloud, "", true)

		StartConfigStreamRoutine(server)
		defer StopConfigStreamRoutine(server)

		time.Sleep(300 * time.Millisecond)
		assert.Equal(t, int32(0), cloud.streamConnections.Load())
	})

	t.Run("it reconnects when the cloud closes the stream", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		cloud.closeStreamImmediately.Store(true)
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go configStreamRoutine(ctx, server, 10, time.Minute)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamConnections.Load() >= 3
		}), "expected the agent to keep reconnecting")
	})

	t.Run("it keeps reconnecting when the cloud fails with a server error", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		cloud.streamStatus.Store(http.StatusInternalServerError)
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go configStreamRoutine(ctx, server, 10, time.Minute)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamConnections.Load() >= 3
		}), "expected a server error to be retried instead of stopping the routine")
	})

	t.Run("it keeps retrying while the cloud is unreachable", func(t *testing.T) {
		cloud := newMockCloud()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		// Close the cloud right away, so connecting is refused
		cloud.httpServer.Close()

		ctx, cancel := context.WithCancel(context.Background())
		routineStopped := make(chan struct{})
		go func() {
			configStreamRoutine(ctx, server, 10, time.Minute)
			close(routineStopped)
		}()

		time.Sleep(300 * time.Millisecond)
		assert.Equal(t, int32(0), cloud.streamConnections.Load())

		// Retrying must not keep the routine from stopping
		cancel()
		select {
		case <-routineStopped:
		case <-time.After(5 * time.Second):
			t.Fatal("expected the config stream routine to stop while the cloud is unreachable")
		}
	})

	t.Run("it drops a connection that stops sending data", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		// The mock keeps the connection open without sending anything after the initial comment
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go configStreamRoutine(ctx, server, 10, 200*time.Millisecond)

		assert.True(t, waitFor(5*time.Second, func() bool {
			return cloud.streamDisconnects.Load() >= 1 && cloud.streamConnections.Load() >= 2
		}), "expected a silent connection to be dropped and reconnected")
	})

	t.Run("it stops for good when the cloud rejects the token", func(t *testing.T) {
		cloud := newMockCloud()
		defer cloud.httpServer.Close()
		cloud.streamStatus.Store(http.StatusUnauthorized)
		server := newTestServerData(cloud, "AIK_RUNTIME_TEST_TOKEN", true)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		routineStopped := make(chan struct{})
		go func() {
			configStreamRoutine(ctx, server, 10, time.Minute)
			close(routineStopped)
		}()

		select {
		case <-routineStopped:
		case <-time.After(5 * time.Second):
			t.Fatal("expected the config stream routine to stop after an authentication failure")
		}

		assert.Equal(t, int32(1), cloud.streamConnections.Load())
	})
}

func TestIsRealtimeEnabled(t *testing.T) {
	newServer := func(sse bool, enabledFeatures []string) *ServerData {
		server := NewServerData()
		server.AikidoConfig.Sse = sse
		server.CloudConfig.EnabledFeatures = enabledFeatures
		return server
	}

	t.Run("it is enabled by the local feature flag", func(t *testing.T) {
		assert.True(t, isRealtimeEnabled(newServer(true, nil)))
	})

	t.Run("it is enabled by the cloud", func(t *testing.T) {
		assert.True(t, isRealtimeEnabled(newServer(false, []string{constants.RealtimeUpdatesFeature})))
		assert.True(t, isRealtimeEnabled(newServer(false, []string{"other_feature", constants.RealtimeUpdatesFeature})))
	})

	t.Run("it is disabled without the local flag and the cloud feature", func(t *testing.T) {
		assert.False(t, isRealtimeEnabled(newServer(false, nil)))
		assert.False(t, isRealtimeEnabled(newServer(false, []string{"other_feature"})))
	})
}

func TestConnectToConfigStream(t *testing.T) {
	t.Run("it parses events terminated with CRLF", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, ": connected\r\n\r\n")
			fmt.Fprint(w, "event: config-updated\r\ndata: {\"configUpdatedAt\":42}\r\n\r\n")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}))
		defer httpServer.Close()

		server := NewServerData()
		server.Logger = log.CreateLogger("test", "ERROR", false)
		server.AikidoConfig.Token = "AIK_RUNTIME_TEST_TOKEN"
		server.AikidoConfig.ConfigEndpoint = httpServer.URL

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		events := make(chan sseEvent, 2)
		go connectToConfigStream(ctx, server, time.Minute, func(event sseEvent) {
			// The keep-alive comment ahead of the event dispatches an empty event, which the handler ignores
			if event.name != "" {
				events <- event
			}
		})

		select {
		case event := <-events:
			assert.Equal(t, "config-updated", event.name)
			assert.Equal(t, `{"configUpdatedAt":42}`, event.data)
		case <-time.After(5 * time.Second):
			t.Fatal("expected an event terminated with CRLF to be parsed")
		}
	})
}
