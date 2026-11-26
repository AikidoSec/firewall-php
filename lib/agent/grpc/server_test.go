package grpc

import (
	"context"
	. "main/aikido_types"
	"main/cloud"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestGetCloudConfigDuringListDownload(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	resume := sync.OnceFunc(func() { close(release) })
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"blockedUserAgents":"new"}`))
	}))
	defer api.Close()
	server := NewServerData()
	server.Logger = log.CreateLogger("test", "ERROR", false)
	server.AikidoConfig = AikidoConfigData{Token: "AIK_TEST", Endpoint: api.URL}
	server.CloudConfig.ConfigUpdatedAt = 1
	server.CloudLists.BlockedUserAgents = "old"
	done := make(chan bool, 1)
	go func() { done <- cloud.StoreCloudConfig(server, []byte(`{"configUpdatedAt":2}`)) }()
	defer resume()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("list download did not start")
	}

	read := make(chan *protos.CloudConfig, 1)
	go func() { read <- getCloudConfig(server, 0) }()
	select {
	case cfg := <-read:
		require.EqualValues(t, 1, cfg.ConfigUpdatedAt)
		require.Equal(t, "old", cfg.BlockedUserAgents)
	case <-time.After(time.Second):
		t.Fatal("cached config read waited for the firewall-list download")
	}
	resume()
	require.True(t, <-done)
	cfg := getCloudConfig(server, 1)
	require.EqualValues(t, 2, cfg.ConfigUpdatedAt)
	require.Equal(t, "new", cfg.BlockedUserAgents)
}

func TestServerSendsLargeCloudConfig(t *testing.T) {
	ips := make([]string, 300000)
	for i := range ips {
		ips[i] = "2001:0db8:0000:0000:0000:0000:0000:0001/128"
	}
	key := ServerKey{Token: "AIK_LARGE_CONFIG_TEST", ServerPID: 1}
	data := &ServerData{
		Logger:      log.CreateLogger("test", "ERROR", false),
		CloudConfig: CloudConfigData{ConfigUpdatedAt: 1},
		CloudLists: CloudListsData{
			BlockedIpsList: map[string]IpBlocklist{
				"geoip/test": {Description: "Geo restrictions", Ips: []string{"198.51.100.0/24"}},
			},
			MonitoredIpsList: map[string]IpBlocklist{"large-list": {Ips: ips}},
		},
	}
	globals.ServersMutex.Lock()
	globals.Servers[key] = data
	globals.ServersMutex.Unlock()
	t.Cleanup(func() {
		globals.ServersMutex.Lock()
		delete(globals.Servers, key)
		globals.ServersMutex.Unlock()
	})
	expected := getCloudConfig(data, 0)
	require.Greater(t, proto.Size(expected), 10*1024*1024)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	transport := newServer()
	go transport.Serve(listener)
	t.Cleanup(transport.Stop)
	connection, err := grpc.Dial(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config, err := protos.NewAikidoClient(connection).GetCloudConfig(ctx, &protos.CloudConfigUpdatedAt{
		Token: key.Token, ServerPid: key.ServerPID,
	})
	require.NoError(t, err)
	require.True(t, proto.Equal(expected, config), "the complete cloud config must reach the worker")
}
