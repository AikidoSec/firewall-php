package cloud

import (
	"fmt"
	. "main/aikido_types"
	"main/log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreCloudConfigRetriesFailedLists(t *testing.T) {
	for _, test := range []struct {
		name              string
		hasPreviousConfig bool
		status            int
		body              string
	}{
		{"startup HTTP error", false, http.StatusServiceUnavailable, "unavailable"},
		{"startup invalid JSON", false, http.StatusOK, "{"},
		{"update HTTP error", true, http.StatusServiceUnavailable, "unavailable"},
		{"update invalid JSON", true, http.StatusOK, "{"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var failLists atomic.Bool
			failLists.Store(true)
			var listRequests atomic.Int32
			config := []byte(`{"success":true,"configUpdatedAt":2,"receivedAnyStats":true}`)
			cloudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/config":
					fmt.Fprint(w, `{"configUpdatedAt":2}`)
				case "/api/runtime/config":
					w.Write(config)
				case "/api/runtime/firewall/lists":
					listRequests.Add(1)
					if failLists.Load() {
						w.WriteHeader(test.status)
						fmt.Fprint(w, test.body)
						return
					}
					fmt.Fprint(w, `{"success":true,"blockedIPAddresses":[{"key":"geoip","description":"Cyprus","ips":["84.46.131.54/32"]}]}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer cloudServer.Close()

			server := NewServerData()
			server.Logger = log.CreateLogger("test", "ERROR", false)
			server.AikidoConfig.Token = "AIK_TEST"
			server.AikidoConfig.Endpoint = cloudServer.URL
			server.AikidoConfig.ConfigEndpoint = cloudServer.URL
			if test.hasPreviousConfig {
				server.CloudConfig = CloudConfigData{
					ConfigUpdatedAt: 1,
					BlockedUserIds:  []string{"blocked-user"},
					BlockedIpsList: map[string]IpBlocklist{
						"previous": {Description: "Existing restriction", Ips: []string{"84.46.131.54/32"}},
					},
				}
			}
			previousConfig := server.CloudConfig

			require.False(t, StoreCloudConfig(server, config))
			require.Equal(t, previousConfig, server.CloudConfig)
			require.True(t, WasConfigUpdated(server, 2))

			failLists.Store(false)
			CheckConfigUpdatedAt(server)
			require.Equal(t, int64(2), server.CloudConfig.ConfigUpdatedAt)
			require.Equal(t, []string{"84.46.131.54/32"}, server.CloudConfig.BlockedIpsList["geoip"].Ips)
			require.Equal(t, int32(2), listRequests.Load())

			CheckConfigUpdatedAt(server)
			require.Equal(t, int32(2), listRequests.Load())
		})
	}
}
