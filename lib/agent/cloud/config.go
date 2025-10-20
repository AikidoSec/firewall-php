package cloud

import (
	"encoding/json"
	. "main/aikido_types"
)

func WasConfigUpdated(server *ServerData, configUpdatedAt int64) bool {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	if configUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		return false
	}
	return true
}

func CheckConfigUpdatedAt(server *ServerData) {
	response, err := SendCloudRequest(server, server.AikidoConfig.ConfigEndpoint, ConfigUpdatedAtAPI, ConfigUpdatedAtMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending polling config request: ", err)
		return
	}

	cloudConfigUpdatedAt := CloudConfigUpdatedAt{}
	err = json.Unmarshal(response, &cloudConfigUpdatedAt)
	if err != nil {
		return
	}

	if !WasConfigUpdated(server, cloudConfigUpdatedAt.ConfigUpdatedAt) {
		return
	}

	configResponse, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, ConfigAPI, ConfigAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending config request: ", err)
		return
	}

	StoreCloudConfig(server, configResponse)
}
