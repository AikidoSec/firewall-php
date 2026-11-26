package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"main/rate_limiting"
	"runtime"
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
	response, err := SendCloudRequest(server, server.AikidoConfig.ConfigEndpoint, constants.ConfigUpdatedAtAPI, constants.ConfigUpdatedAtMethod, nil)
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

	FetchAndStoreCloudConfig(server)
}

func FetchAndStoreCloudConfig(server *ServerData) {
	configResponse, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.ConfigAPI, constants.ConfigAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending config request: ", err)
		return
	}

	StoreCloudConfig(server, configResponse)
}

func StoreCloudConfig(server *ServerData, configReponse []byte) bool {
	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(configReponse, &tempCloudConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal cloud config: %v", err)
		return false
	}
	server.ConfigUpdateMutex.Lock()
	defer server.ConfigUpdateMutex.Unlock()
	if tempCloudConfig.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		log.Debugf(server.Logger, "ConfigUpdatedAt is the same!")
		return true
	}
	lists := fetchCloudLists(server)
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	server.CloudConfig = tempCloudConfig
	if lists != nil {
		server.CloudLists = *lists
	}
	ApplyCloudConfig(server)
	return true
}

func ApplyCloudConfig(server *ServerData) {
	log.Infof(server.Logger, "Applying new cloud config: %v", server.CloudConfig)
	ResetHeartbeatTicker(server)
	rate_limiting.UpdateConfig(server)
}

func fetchCloudLists(server *ServerData) *CloudListsData {
	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.ListsAPI, constants.ListsAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending lists request: ", err)
		return nil
	}

	tempListsConfig := ListsConfigData{}
	err = json.Unmarshal(response, &tempListsConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal lists config: %v", err)
		return nil
	}

	lists := &CloudListsData{
		BlockedIpsList:      UpdateIpsLists(tempListsConfig.BlockedIpAddresses),
		MonitoredIpsList:    UpdateIpsLists(tempListsConfig.MonitoredIpAddresses),
		AllowedIpsList:      UpdateIpsLists(tempListsConfig.AllowedIpAddresses),
		BlockedUserAgents:   tempListsConfig.BlockedUserAgents,
		MonitoredUserAgents: tempListsConfig.MonitoredUserAgents,
		UserAgentDetails:    make(map[string]string, len(tempListsConfig.UserAgentDetails)),
	}
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		lists.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}

	/* Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP */
	tempListsConfig = ListsConfigData{}
	runtime.GC()

	return lists
}

func UpdateIpsLists(ipLists []IpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, ipList := range ipLists {
		m[ipList.Key] = IpBlocklist{Description: ipList.Description, Ips: ipList.Ips}
	}
	return m
}
