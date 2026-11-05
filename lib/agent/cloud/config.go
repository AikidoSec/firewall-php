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
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	tempCloudConfig := CloudConfigData{}
	err := json.Unmarshal(configReponse, &tempCloudConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal cloud config: %v", err)
		return false
	}
	if tempCloudConfig.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		log.Debugf(server.Logger, "ConfigUpdatedAt is the same!")
		return true
	}
	server.CloudConfig = tempCloudConfig
	UpdateListsConfig(server)
	ApplyCloudConfig(server)
	return true
}

func ApplyCloudConfig(server *ServerData) {
	log.Infof(server.Logger, "Applying new cloud config: %v", server.CloudConfig)
	ResetHeartbeatTicker(server)
	rate_limiting.UpdateConfig(server)
}

func UpdateListsConfig(server *ServerData) bool {
	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.ListsAPI, constants.ListsAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending lists request: ", err)
		return false
	}

	tempListsConfig := ListsConfigData{}
	err = json.Unmarshal(response, &tempListsConfig)
	if err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal lists config: %v", err)
		return false
	}

	server.CloudLists.BlockedIpsList = UpdateIpsLists(tempListsConfig.BlockedIpAddresses)
	server.CloudLists.MonitoredIpsList = UpdateIpsLists(tempListsConfig.MonitoredIpAddresses)
	server.CloudLists.AllowedIpsList = UpdateIpsLists(tempListsConfig.AllowedIpAddresses)

	server.CloudLists.BlockedUserAgents = tempListsConfig.BlockedUserAgents
	server.CloudLists.MonitoredUserAgents = tempListsConfig.MonitoredUserAgents

	server.CloudLists.UserAgentDetails = make(map[string]string)
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		server.CloudLists.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}

	/* Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP */
	tempListsConfig = ListsConfigData{}
	runtime.GC()

	return true
}

func UpdateIpsLists(ipLists []IpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, ipList := range ipLists {
		m[ipList.Key] = IpBlocklist{Description: ipList.Description, Ips: ipList.Ips}
	}
	return m
}
