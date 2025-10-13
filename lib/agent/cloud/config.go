package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
)

func CheckConfigUpdatedAt() {
	for _, server := range globals.Servers {
		response, err := SendCloudRequest(server, server.EnvironmentConfig.ConfigEndpoint, globals.ConfigUpdatedAtAPI, globals.ConfigUpdatedAtMethod, nil)
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

		configResponse, err := SendCloudRequest(server, server.EnvironmentConfig.Endpoint, globals.ConfigAPI, globals.ConfigAPIMethod, nil)
		if err != nil {
			LogCloudRequestError(server, "Error in sending config request: ", err)
			return
		}

		StoreCloudConfig(server, configResponse)
	}
}
