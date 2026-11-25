package server_utils

import (
	"main/aikido_types"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

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

	release()
	select {
	case <-registerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("registration did not complete after the cloud request was released")
	}
}
