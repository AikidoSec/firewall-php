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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestGetCloudConfigSharesDownloadsAcrossServerPIDs(t *testing.T) {
	log.Init()
	t.Cleanup(log.Uninit)
	var polls, downloads, lists atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case constants.ConfigUpdatedAtAPI:
			fmt.Fprintf(w, `{"configUpdatedAt":%d}`, polls.Add(1))
		case constants.ConfigAPI:
			fmt.Fprintf(w, `{"configUpdatedAt":%d,"endpoints":[{"method":"GET","route":"/pets","rateLimiting":{"enabled":true,"windowSizeInMS":60000,"maxRequests":10}}]}`, downloads.Add(1))
		case "/" + constants.ListsAPI:
			fmt.Fprintf(w, `{"blockedUserAgents":"bot-%d"}`, lists.Add(1))
		default:
			fmt.Fprint(w, "{}")
		}
	}))
	t.Cleanup(api.Close)
	handler := &GrpcServer{}
	const token = "AIK_SHARED_POLICY"
	register := func(pid int32, token, endpoint, configEndpoint string) {
		key := ServerKey{Token: token, ServerPID: pid}
		_, err := handler.OnConfig(t.Context(), &protos.Config{Token: token, ServerPid: pid, Endpoint: endpoint, ConfigEndpoint: configEndpoint})
		if err != nil {
			t.Errorf("registration failed: %v", err)
		}
		t.Cleanup(func() {
			server_utils.Unregister(key)
			globals.ServersMutex.Lock()
			delete(globals.PastDeletedServers, key)
			globals.ServersMutex.Unlock()
		})
	}
	var registrations sync.WaitGroup
	for _, pid := range []int32{1, 2} {
		registrations.Go(func() { register(pid, token, api.URL, api.URL) })
	}
	registrations.Wait()
	assertRevision := func(version int32) {
		t.Helper()
		for _, pid := range []int32{1, 2} {
			cfg, err := handler.GetCloudConfig(t.Context(), &protos.CloudConfigUpdatedAt{Token: token, ServerPid: pid, ConfigUpdatedAt: int64(version - 1)})
			require.NoError(t, err)
			require.EqualValues(t, version, cfg.ConfigUpdatedAt)
			require.Equal(t, fmt.Sprintf("bot-%d", version), cfg.BlockedUserAgents)
			_, err = handler.GetCloudConfig(t.Context(), &protos.CloudConfigUpdatedAt{Token: token, ServerPid: pid, ConfigUpdatedAt: cfg.ConfigUpdatedAt})
			require.Error(t, err)
		}
		require.Equal(t, version, polls.Load())
		require.Equal(t, version, downloads.Load())
		require.Equal(t, version, lists.Load())
	}
	assertRevision(1)
	a := globals.GetServer(ServerKey{Token: token, ServerPID: 1})
	b := globals.GetServer(ServerKey{Token: token, ServerPID: 2})
	a.ConfigRefreshMutex.Lock()
	registered := make(chan struct{})
	go func() {
		register(6, token, api.URL, api.URL)
		close(registered)
	}()
	select {
	case <-registered:
	case <-time.After(time.Second):
		t.Error("registration with cached config waited for a refresh")
	}
	a.ConfigRefreshMutex.Unlock()
	<-registered
	cloud.PollCloudConfigs(nil)
	assertRevision(2)
	route := RateLimitingKey{Method: "GET", Route: "/pets"}
	require.NotNil(t, a.RateLimitingMap[route])
	require.NotNil(t, b.RateLimitingMap[route])
	require.NotSame(t, a.RateLimitingMap[route], b.RateLimitingMap[route])
	require.True(t, cloud.UpdateCloudConfig(a, []byte(`{"configUpdatedAt":3}`)))
	cfg, err := handler.GetCloudConfig(t.Context(), &protos.CloudConfigUpdatedAt{Token: token, ServerPid: 2, ConfigUpdatedAt: 2})
	require.NoError(t, err)
	require.EqualValues(t, 3, cfg.ConfigUpdatedAt)
	require.Equal(t, "bot-3", cfg.BlockedUserAgents)

	otherAPI := httptest.NewServer(api.Config.Handler)
	t.Cleanup(otherAPI.Close)
	register(3, "AIK_OTHER", api.URL, api.URL)
	register(4, token, otherAPI.URL, api.URL)
	register(5, token, api.URL, otherAPI.URL)
	require.EqualValues(t, 5, downloads.Load())
}

func TestGetCloudConfigDuringListDownload(t *testing.T) {
	server := NewServerData()
	server.Logger = log.CreateLogger("test", "ERROR", false)
	server.CloudConfig.ConfigUpdatedAt = 1
	server.CloudLists.BlockedUserAgents = "old"

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
	require.True(t, cloud.UpdateCloudConfig(server, []byte(`{"configUpdatedAt":2}`)))
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
	data.CloudConfig.ConfigUpdatedAt = 1
	data.CloudLists = CloudListsData{
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
