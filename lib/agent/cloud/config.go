package cloud

import (
	"encoding/json"
	"main/aikido_types"
	. "main/aikido_types"
)

func CheckConfigUpdatedAt(server *ServerData) {
	response, err := SendCloudRequest(server, server.AikidoConfig.ConfigEndpoint, aikido_types.ConfigUpdatedAtAPI, aikido_types.ConfigUpdatedAtMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending polling config request: ", err)
		return
	}

	cloudConfigUpdatedAt := CloudConfigUpdatedAt{}
	err = json.Unmarshal(response, &cloudConfigUpdatedAt)
	if err != nil {
		return
	}

	if cloudConfigUpdatedAt.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		return
	}

	configResponse, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, ConfigAPI, ConfigAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending config request: ", err)
		return
	}

	StoreCloudConfig(server, configResponse)
}
