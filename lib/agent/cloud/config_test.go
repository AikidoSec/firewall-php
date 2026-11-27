package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPollKeepsCachedConfigWhenDownloadedRevisionIsOlder(t *testing.T) {
	var downloaded atomic.Bool
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case constants.ConfigUpdatedAtAPI:
			_, _ = w.Write([]byte(`{"configUpdatedAt":3}`))
		case constants.ConfigAPI:
			downloaded.Store(true)
			_, _ = w.Write([]byte(`{"configUpdatedAt":1}`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	server := NewServerData()
	server.Logger = log.CreateLogger("test", "ERROR", false)
	server.AikidoConfig = AikidoConfigData{Token: "AIK_TEST", Endpoint: api.URL, ConfigEndpoint: api.URL}
	server.CloudPolicy.CloudConfig.ConfigUpdatedAt = 2
	server.CloudPolicy.CloudLists.BlockedUserAgents = "cached"

	CheckConfigUpdatedAt(server)

	require.True(t, downloaded.Load())
	require.EqualValues(t, 2, server.CloudPolicy.CloudConfig.ConfigUpdatedAt)
	require.Equal(t, "cached", server.CloudPolicy.CloudLists.BlockedUserAgents)
}

func TestStoreCloudConfigKeepsListsOnFailedDownload(t *testing.T) {
	cloudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer cloudServer.Close()
	heartbeat := time.NewTicker(time.Hour)
	defer heartbeat.Stop()
	server := &ServerData{
		Logger:       log.CreateLogger("test", "ERROR", false),
		AikidoConfig: AikidoConfigData{Token: "AIK_TEST", Endpoint: cloudServer.URL},
		PollingData:  &ServerDataPolling{HeartbeatTicker: heartbeat},
		CloudPolicy:  &CloudPolicy{CloudConfig: CloudConfigData{ConfigUpdatedAt: 1}},
	}
	server.CloudPolicy.CloudLists = CloudListsData{
		BlockedIpsList:      map[string]IpBlocklist{"blocked": {Ips: []string{"1.2.3.4"}}},
		AllowedIpsList:      map[string]IpBlocklist{"allowed": {Ips: []string{"1.2.3.5"}}},
		MonitoredIpsList:    map[string]IpBlocklist{"monitored": {Ips: []string{"1.2.3.6"}}},
		BlockedUserAgents:   "BlockedBot",
		MonitoredUserAgents: "MonitoredBot",
		UserAgentDetails:    map[string]string{"bot": "BlockedBot|MonitoredBot"},
	}
	lists := server.CloudPolicy.CloudLists
	require.True(t, StoreCloudConfig(server, []byte(`{"configUpdatedAt":2,"block":true}`)))
	require.Equal(t, int64(2), server.CloudPolicy.CloudConfig.ConfigUpdatedAt)
	require.NotNil(t, server.CloudPolicy.CloudConfig.Block)
	require.True(t, *server.CloudPolicy.CloudConfig.Block)
	require.Equal(t, lists, server.CloudPolicy.CloudLists)
}
