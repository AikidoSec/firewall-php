package config

import (
	"main/aikido_types"
	"main/globals"
	"main/instance"
	"testing"
)

func TestInitUsesServerPIDFromExtension(t *testing.T) {
	previousConfig := globals.EnvironmentConfig
	t.Cleanup(func() {
		globals.EnvironmentConfig = previousConfig
	})

	const platformName = "test-sapi"
	const serverPID int32 = 1234
	Init(platformName, serverPID)

	if globals.EnvironmentConfig.ServerPID != serverPID {
		t.Fatalf("ServerPID = %d, expected %d", globals.EnvironmentConfig.ServerPID, serverPID)
	}
}

func TestReloadClearsTokenlessSiteAndRestoresCachedServer(t *testing.T) {
	previousServers := globals.Servers
	globals.Servers = make(map[string]*aikido_types.ServerData)
	t.Cleanup(func() { globals.Servers = previousServers })

	processor := instance.NewRequestProcessorInstance(1)
	var siteA *aikido_types.ServerData
	for _, token := range []string{"site-a", "site-a", "site-b", "", "", "site-a"} {
		configJson := `{"token":"` + token + `","log_level":"WARN"}`
		conf := aikido_types.AikidoConfigData{}
		if !ReloadAikidoConfig(processor, &conf, configJson) {
			t.Fatalf("token %q: reload failed", token)
		}
		if processor.GetCurrentToken() != token {
			t.Fatalf("token %q: retained token %q", token, processor.GetCurrentToken())
		}
		server := processor.GetCurrentServer()
		if token == "" {
			if server != nil || processor.IsInitialized() {
				t.Fatal("tokenless site retained the previous server")
			}
		} else if server == nil || server.AikidoConfig.Token != token {
			t.Fatalf("token %q: selected the wrong server", token)
		}
		if token == "site-a" {
			if siteA != nil && server != siteA {
				t.Fatal("returning to site A did not reuse its cached server")
			}
			siteA = server
		}
	}
}
