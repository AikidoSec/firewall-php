package grpc

import (
	"context"
	"encoding/json"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestOnAttackDetectedReturnsBeforeCloudUploadCompletes(t *testing.T) {
	releaseUpload := make(chan struct{})
	receivedEvent := make(chan DetectedAttack, 1)
	cloudServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != constants.EventsAPI {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		var event DetectedAttack
		if err := json.NewDecoder(request.Body).Decode(&event); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		receivedEvent <- event
		<-releaseUpload
		_, _ = writer.Write([]byte("{}"))
	}))
	defer cloudServer.Close()
	defer close(releaseUpload)

	key := ServerKey{Token: "AIK_ATTACK_EVENT_TEST", ServerPID: 1}
	server := &ServerData{
		AikidoConfig: AikidoConfigData{Token: key.Token, Endpoint: cloudServer.URL},
		Logger:       log.CreateLogger("test", "ERROR", false),
	}
	globals.ServersMutex.Lock()
	globals.Servers[key] = server
	globals.ServersMutex.Unlock()
	t.Cleanup(func() {
		globals.ServersMutex.Lock()
		delete(globals.Servers, key)
		globals.ServersMutex.Unlock()
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	transport := newServer()
	go transport.Serve(listener)
	t.Cleanup(transport.Stop)
	connection, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = protos.NewAikidoClient(connection).OnAttackDetected(ctx, &protos.AttackDetected{
		Token: key.Token, ServerPid: key.ServerPID,
		Request: &protos.Request{Route: "/callback"},
		Attack:  &protos.Attack{Kind: "ssrf", Blocked: true},
	})
	require.NoError(t, err, "the RPC must return while the cloud response is held open")
	server.StatsData.StatsMutex.Lock()
	attacks, blocked := server.StatsData.Attacks, server.StatsData.AttacksBlocked
	server.StatsData.StatsMutex.Unlock()
	require.Equal(t, 1, attacks)
	require.Equal(t, 1, blocked)

	select {
	case event := <-receivedEvent:
		require.Equal(t, "detected_attack", event.Type)
		require.Equal(t, "/callback", event.Request.Route)
		require.Equal(t, "ssrf", event.Attack.Kind)
		require.True(t, event.Attack.Blocked)
	case <-ctx.Done():
		t.Fatal("expected the attack event to reach the cloud")
	}
}
