package cloud

import (
	. "main/aikido_types"
	"main/log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
		CloudConfig:  CloudConfigData{ConfigUpdatedAt: 1},
		PollingData:  &ServerDataPolling{HeartbeatTicker: heartbeat},
		CloudLists: CloudListsData{
			BlockedIpsList:      map[string]IpBlocklist{"blocked": {Ips: []string{"1.2.3.4"}}},
			AllowedIpsList:      map[string]IpBlocklist{"allowed": {Ips: []string{"1.2.3.5"}}},
			MonitoredIpsList:    map[string]IpBlocklist{"monitored": {Ips: []string{"1.2.3.6"}}},
			BlockedUserAgents:   "BlockedBot",
			MonitoredUserAgents: "MonitoredBot",
			UserAgentDetails:    map[string]string{"bot": "BlockedBot|MonitoredBot"},
		},
	}
	lists := server.CloudLists
	require.True(t, StoreCloudConfig(server, []byte(`{"configUpdatedAt":2,"block":true}`)))
	require.Equal(t, int64(2), server.CloudConfig.ConfigUpdatedAt)
	require.NotNil(t, server.CloudConfig.Block)
	require.True(t, *server.CloudConfig.Block)
	require.Equal(t, lists, server.CloudLists)
}
