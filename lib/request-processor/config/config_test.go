package config

import (
	"main/globals"
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
