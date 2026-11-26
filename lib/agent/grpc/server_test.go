package grpc

import (
	"context"
	"fmt"
	. "main/aikido_types"
	"main/cloud"
	"main/constants"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/server_utils"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestRegistrationReusesCachedCloudPolicy(t *testing.T) {
	log.Init()
	t.Cleanup(log.Uninit)
	var polls, downloads, lists atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case constants.ConfigUpdatedAtAPI:
			polls.Add(1)
			fmt.Fprint(w, `{"configUpdatedAt":1}`)
		case constants.ConfigAPI:
			downloads.Add(1)
			fmt.Fprint(w, `{"configUpdatedAt":1,"endpoints":[{"method":"GET","route":"/pets","rateLimiting":{"enabled":true,"windowSizeInMS":60000,"maxRequests":10}}]}`)
		case "/" + constants.ListsAPI:
			fmt.Fprintf(w, `{"blockedUserAgents":"bot-%d"}`, lists.Add(1))
		default:
			fmt.Fprint(w, "{}")
		}
	}))
	t.Cleanup(api.Close)
	handler := &GrpcServer{}
	first := ServerKey{Token: "AIK_SHARED_POLICY", ServerPID: 1}
	second := ServerKey{Token: first.Token, ServerPID: 2}
	register := func(key ServerKey, endpoint, configEndpoint string, blocking bool) *ServerData {
		t.Helper()
		_, err := handler.OnConfig(t.Context(), &protos.Config{
			Token: key.Token, ServerPid: key.ServerPID, Endpoint: endpoint, ConfigEndpoint: configEndpoint, Blocking: blocking,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			server_utils.Unregister(key)
			globals.ServersMutex.Lock()
			delete(globals.PastDeletedServers, key)
			globals.ServersMutex.Unlock()
		})
		return globals.GetServer(key)
	}
	a := register(first, api.URL, api.URL, false)
	b := register(second, api.URL, api.URL, true)
	require.NotSame(t, a, b)
	require.Same(t, a.CloudPolicy, b.CloudPolicy)
	require.EqualValues(t, 1, polls.Load())
	require.EqualValues(t, 1, downloads.Load())
	require.EqualValues(t, 1, lists.Load())
	require.Equal(t, "bot-1", getCloudConfig(b, 0).GetBlockedUserAgents())
	require.False(t, getCloudConfig(a, 0).GetBlock())
	require.True(t, getCloudConfig(b, 0).GetBlock())

	route := RateLimitingKey{Method: "GET", Route: "/pets"}
	require.NotNil(t, a.RateLimitingMap[route])
	require.NotNil(t, b.RateLimitingMap[route])
	require.NotSame(t, a.RateLimitingMap[route], b.RateLimitingMap[route])
	require.True(t, cloud.StoreCloudConfig(a, []byte(`{"configUpdatedAt":2}`)))
	require.Equal(t, "bot-2", getCloudConfig(b, 1).GetBlockedUserAgents())
	require.NotContains(t, b.RateLimitingMap, route)

	otherAPI := httptest.NewServer(api.Config.Handler)
	t.Cleanup(otherAPI.Close)
	require.NotSame(t, a.CloudPolicy, register(ServerKey{Token: "AIK_OTHER_POLICY", ServerPID: first.ServerPID}, api.URL, api.URL, false).CloudPolicy)
	require.NotSame(t, a.CloudPolicy, register(ServerKey{Token: first.Token, ServerPID: 4}, otherAPI.URL, api.URL, false).CloudPolicy)
	require.NotSame(t, a.CloudPolicy, register(ServerKey{Token: first.Token, ServerPID: 5}, api.URL, otherAPI.URL, false).CloudPolicy)
}

func TestGetCloudConfigDuringListDownload(t *testing.T) {
	server := NewServerData()
	server.Logger = log.CreateLogger("test", "ERROR", false)
	server.CloudPolicy.CloudConfig.ConfigUpdatedAt = 1
	server.CloudPolicy.CloudLists.BlockedUserAgents = "old"

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		read := make(chan *protos.CloudConfig, 1)
		go func() { read <- getCloudConfig(server, 0) }()
		select {
		case cfg := <-read:
			if cfg.GetConfigUpdatedAt() != 1 || cfg.GetBlockedUserAgents() != "old" {
				t.Errorf("during download got revision %d and list %q", cfg.GetConfigUpdatedAt(), cfg.GetBlockedUserAgents())
			}
		case <-time.After(time.Second):
			t.Error("cached config read waited for the firewall-list response")
		}
		_, _ = w.Write([]byte(`{"blockedUserAgents":"new"}`))
	}))
	defer api.Close()
	server.AikidoConfig = AikidoConfigData{Token: "AIK_TEST", Endpoint: api.URL}
	require.True(t, cloud.StoreCloudConfig(server, []byte(`{"configUpdatedAt":2}`)))
	cfg := getCloudConfig(server, 1)
	require.EqualValues(t, 2, cfg.GetConfigUpdatedAt())
	require.Equal(t, "new", cfg.GetBlockedUserAgents())
}

func TestServerSendsLargeCloudConfig(t *testing.T) {
	ips := make([]string, 300000)
	for i := range ips {
		ips[i] = "2001:0db8:0000:0000:0000:0000:0000:0001/128"
	}
	key := ServerKey{Token: "AIK_LARGE_CONFIG_TEST", ServerPID: 1}
	data := NewServerData()
	data.Logger = log.CreateLogger("test", "ERROR", false)
	data.CloudPolicy.CloudConfig.ConfigUpdatedAt = 1
	data.CloudPolicy.CloudLists = CloudListsData{
		BlockedIpsList: map[string]IpBlocklist{
			"geoip/test": {Description: "Geo restrictions", Ips: []string{"198.51.100.0/24"}},
		},
		MonitoredIpsList: map[string]IpBlocklist{"large-list": {Ips: ips}},
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
