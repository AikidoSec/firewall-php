package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/log"
	"main/rate_limiting"
	"runtime"
)

func WasConfigUpdated(server *ServerData, configUpdatedAt int64) bool {
	server.CloudPolicy.ConfigCacheMutex.Lock()
	defer server.CloudPolicy.ConfigCacheMutex.Unlock()
	if configUpdatedAt <= server.CloudPolicy.CloudConfig.ConfigUpdatedAt {
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
	server.CloudPolicy.ConfigRefreshMutex.Lock()
	defer server.CloudPolicy.ConfigRefreshMutex.Unlock()

	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(configReponse, &tempCloudConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal cloud config: %v", err)
		return false
	}
	if !WasConfigUpdated(server, tempCloudConfig.ConfigUpdatedAt) {
		log.Debugf(server.Logger, "ConfigUpdatedAt is the same!")
		return true
	}
	lists := fetchCloudLists(server)
	ApplyCloudConfig(server, tempCloudConfig, lists)
	if lists != nil {
		// A large list refresh can leave old lists and temporary buffers on the heap until the next GC cycle.
		runtime.GC()
	}
	return true
}

func ApplyCloudConfig(server *ServerData, config CloudConfigData, lists *ListsConfigData) {
	policy := server.CloudPolicy
	policy.ConfigCacheMutex.Lock()
	defer policy.ConfigCacheMutex.Unlock()
	policy.CloudConfig = config
	if lists != nil {
		UpdateListsConfig(server, lists)
	}
	updateServerSettings(server)
	for _, other := range globals.GetServers() {
		if other != server && other.CloudPolicy == policy {
			updateServerSettings(other)
		}
	}
}

func updateServerSettings(server *ServerData) bool {
	if server.CloudPolicy.CloudConfig.ConfigUpdatedAt == 0 {
		return false
	}
	log.Infof(server.Logger, "Applying new cloud config: %v", server.CloudPolicy.CloudConfig)
	ResetHeartbeatTicker(server)
	rate_limiting.UpdateConfig(server)
	return true
}

func fetchCloudLists(server *ServerData) *ListsConfigData {
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

	return &tempListsConfig
}

func UpdateListsConfig(server *ServerData, tempListsConfig *ListsConfigData) {
	server.CloudPolicy.CloudLists.BlockedIpsList = UpdateIpsLists(tempListsConfig.BlockedIpAddresses)
	server.CloudPolicy.CloudLists.MonitoredIpsList = UpdateIpsLists(tempListsConfig.MonitoredIpAddresses)
	server.CloudPolicy.CloudLists.AllowedIpsList = UpdateIpsLists(tempListsConfig.AllowedIpAddresses)

	server.CloudPolicy.CloudLists.BlockedUserAgents = tempListsConfig.BlockedUserAgents
	server.CloudPolicy.CloudLists.MonitoredUserAgents = tempListsConfig.MonitoredUserAgents

	server.CloudPolicy.CloudLists.UserAgentDetails = make(map[string]string)
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		server.CloudPolicy.CloudLists.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}
}

func UpdateIpsLists(ipLists []IpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, ipList := range ipLists {
		m[ipList.Key] = IpBlocklist{Description: ipList.Description, Ips: ipList.Ips}
	}
	return m
}
