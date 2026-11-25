package server_utils

import (
	"fmt"
	"main/aikido_types"
	"main/constants"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServersShareCloudPolicy(t *testing.T) {
	log.Init()
	t.Cleanup(log.Uninit)
	var polls, configs, lists, starts atomic.Int32
	var revision atomic.Int64
	revision.Store(1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case constants.ConfigUpdatedAtAPI:
			polls.Add(1)
			fmt.Fprintf(w, `{"configUpdatedAt":%d}`, revision.Load())
		case constants.ConfigAPI:
			configs.Add(1)
			fmt.Fprintf(w, `{"configUpdatedAt":%d,"endpoints":[{"method":"GET","route":"/pets","rateLimiting":{"enabled":true,"windowSizeInMS":60000,"maxRequests":10}}]}`, revision.Load())
		case "/" + constants.ListsAPI:
			lists.Add(1)
			fmt.Fprint(w, `{"blockedUserAgents":"test-bot"}`)
		case constants.EventsAPI:
			starts.Add(1)
			fmt.Fprint(w, "{}")
		}
	})
	api := httptest.NewServer(handler)
	t.Cleanup(api.Close)
	otherAPI := httptest.NewServer(handler)
	t.Cleanup(otherAPI.Close)

	register := func(pid int32, token, endpoint, configEndpoint string) *aikido_types.ServerData {
		key := aikido_types.ServerKey{Token: token, ServerPID: pid}
		Register(key, pid, &protos.Config{Token: token, Endpoint: endpoint, ConfigEndpoint: configEndpoint, LogLevel: "ERROR"})
		t.Cleanup(func() {
			if globals.GetServer(key) != nil {
				Unregister(key)
			}
			globals.ServersMutex.Lock()
			delete(globals.PastDeletedServers, key)
			globals.ServersMutex.Unlock()
		})
		return globals.GetServer(key)
	}
	const token = "AIK_SHARED_CONFIG"
	var first, second *aikido_types.ServerData
	var wg sync.WaitGroup
	wg.Go(func() { first = register(1, token, api.URL, api.URL) })
	wg.Go(func() { second = register(2, token, api.URL, api.URL) })
	wg.Wait()
	require.Same(t, first.CloudPolicy, second.CloudPolicy)
	require.EqualValues(t, 1, polls.Load())
	require.EqualValues(t, 1, configs.Load())
	require.EqualValues(t, 1, lists.Load())
	require.EqualValues(t, 2, starts.Load())
	require.Equal(t, "test-bot", second.CloudLists.BlockedUserAgents)
	route := aikido_types.RateLimitingKey{Method: "GET", Route: "/pets"}
	require.NotNil(t, first.RateLimitingMap[route])
	require.NotNil(t, second.RateLimitingMap[route])
	require.NotSame(t, first.RateLimitingMap[route], second.RateLimitingMap[route])

	require.NotSame(t, first.CloudPolicy, register(3, "AIK_OTHER", api.URL, api.URL).CloudPolicy)
	require.NotSame(t, first.CloudPolicy, register(4, token, otherAPI.URL, api.URL).CloudPolicy)
	require.NotSame(t, first.CloudPolicy, register(5, token, api.URL, otherAPI.URL).CloudPolicy)

	ownerPID, remaining := int32(1), second
	if first.Servers[0] == second {
		ownerPID, remaining = 2, first
	}
	Unregister(aikido_types.ServerKey{Token: token, ServerPID: ownerPID})
	revision.Store(2)
	remaining.PollingData.ConfigPollingTicker.Reset(10 * time.Millisecond)
	require.Eventually(t, func() bool {
		remaining.CloudConfigMutex.Lock()
		defer remaining.CloudConfigMutex.Unlock()
		return remaining.CloudConfig.ConfigUpdatedAt == 2
	}, 2*time.Second, 5*time.Millisecond)
}

func TestRegisterDoesNotBlockServerLookupsDuringCloudRequest(t *testing.T) {
	log.Init()
	t.Cleanup(log.Uninit)

	cloudRequestStarted := make(chan struct{})
	releaseCloudRequest := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseCloudRequest) }) }

	cloudServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/config" {
			select {
			case <-cloudRequestStarted:
			default:
				close(cloudRequestStarted)
			}
			<-releaseCloudRequest
			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{"configUpdatedAt":0}`))
			return
		}

		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{}`))
	}))

	serverKey := aikido_types.ServerKey{Token: "AIK_RUNTIME_TEST", ServerPID: 12345}
	request := &protos.Config{
		Token:          serverKey.Token,
		ServerPid:      serverKey.ServerPID,
		Endpoint:       cloudServer.URL,
		ConfigEndpoint: cloudServer.URL,
		LogLevel:       "ERROR",
	}

	registerDone := make(chan struct{})
	go func() {
		Register(serverKey, 54321, request)
		close(registerDone)
	}()

	t.Cleanup(func() {
		release()
		select {
		case <-registerDone:
		case <-time.After(2 * time.Second):
		}
		if globals.GetServer(serverKey) != nil {
			Unregister(serverKey)
		}
		globals.ServersMutex.Lock()
		delete(globals.PastDeletedServers, serverKey)
		globals.ServersMutex.Unlock()
		cloudServer.Close()
	})

	select {
	case <-cloudRequestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("registration did not reach the cloud request")
	}

	lookupDone := make(chan *aikido_types.ServerData, 1)
	go func() { lookupDone <- globals.GetServer(serverKey) }()

	select {
	case server := <-lookupDone:
		if server == nil {
			t.Fatal("registered server was not visible during the cloud request")
		}
		if server.StatsData.StartedAt == 0 || server.StatsData.MonitoredSinkTimings == nil {
			t.Fatal("registered server was visible before its local state was initialized")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("server lookup was blocked by the cloud request")
	}

	select {
	case <-registerDone:
		t.Fatal("registration completed before the stalled cloud request was released")
	default:
	}

	server := globals.GetServer(serverKey)
	unregistered := make(chan struct{})
	go func() {
		Unregister(serverKey)
		close(unregistered)
	}()
	require.Eventually(t, func() bool { return globals.GetServer(serverKey) == nil }, time.Second, time.Millisecond)
	release()
	select {
	case <-unregistered:
		require.Empty(t, server.CloudPolicy.Servers)
	case <-time.After(2 * time.Second):
		t.Fatal("unregistration did not complete after the cloud request was released")
	}
}
