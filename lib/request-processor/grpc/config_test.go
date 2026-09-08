package grpc

import (
	"main/globals"
	"main/ipc/protos"
	"testing"
)

func TestSetCloudConfigIgnoresOlderOrRepeatedVersions(t *testing.T) {
	server := globals.NewServerData()
	setCloudConfig(server, &protos.CloudConfig{ConfigUpdatedAt: 2, Block: true})

	for _, version := range []int64{1, 2} {
		setCloudConfig(server, &protos.CloudConfig{ConfigUpdatedAt: version, Block: false})
		if server.CloudConfig.ConfigUpdatedAt != 2 || server.CloudConfig.Block != 1 {
			t.Fatalf("version %d replaced the current blocking policy", version)
		}
	}

	setCloudConfig(server, &protos.CloudConfig{ConfigUpdatedAt: 3, Block: false})
	if server.CloudConfig.ConfigUpdatedAt != 3 || server.CloudConfig.Block != 0 {
		t.Fatal("newer cloud config was not applied")
	}
}
