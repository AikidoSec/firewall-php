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

func announcesNewerConfig(server *ServerData, payload []byte) bool {
	var update CloudConfigUpdatedAt
	if err := json.Unmarshal(payload, &update); err != nil {
		log.Debugf(server.Logger, "Invalid config update notification: %v", err)
		return false
	}
	return update.ConfigUpdatedAt > server.CloudConfig.ConfigUpdatedAt
}

func CheckConfigUpdatedAt(server *ServerData) {
	server.ConfigUpdateMutex.Lock()
	defer server.ConfigUpdateMutex.Unlock()
	checkConfigUpdatedAt(server)
}

func checkConfigUpdatedAt(server *ServerData) {
	response, err := SendCloudRequest(server, server.AikidoConfig.ConfigEndpoint, constants.ConfigUpdatedAtAPI, constants.ConfigUpdatedAtMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending polling config request: ", err)
		return
	}

	if announcesNewerConfig(server, response) {
		fetchAndStoreCloudConfig(server)
	}
}

func fetchAndStoreCloudConfig(server *ServerData) {
	configResponse, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.ConfigAPI, constants.ConfigAPIMethod, nil)
	if err != nil {
		LogCloudRequestError(server, "Error in sending config request: ", err)
		return
	}

	if config := parseCloudConfig(server, configResponse); config != nil {
		storeCloudConfig(server, config)
	}
}

func parseCloudConfig(server *ServerData, response []byte) *CloudConfigData {
	var config CloudConfigData
	if err := json.Unmarshal(response, &config); err != nil {
		log.Warnf(server.Logger, "Failed to unmarshal cloud config: %v", err)
		return nil
	}
	return &config
}

func StoreCloudConfig(server *ServerData, response []byte) bool {
	config := parseCloudConfig(server, response)
	if config == nil {
		return false
	}
	server.CloudConfigMutex.Lock()
	alreadyStored := config.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt
	server.CloudConfigMutex.Unlock()
	if alreadyStored {
		return true
	}
	server.ConfigUpdateMutex.Lock()
	defer server.ConfigUpdateMutex.Unlock()
	storeCloudConfig(server, config)
	return true
}

func storeCloudConfig(server *ServerData, config *CloudConfigData) {
	if config.ConfigUpdatedAt <= server.CloudConfig.ConfigUpdatedAt {
		return
	}
	lists := fetchCloudLists(server)
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()
	server.CloudConfig = *config
	if lists != nil {
		server.CloudLists = *lists
	}
	ApplyCloudConfig(server)
	for _, other := range globals.GetServers() {
		if other != server && other.CloudPolicy == server.CloudPolicy {
			ApplyCloudConfig(other)
		}
	}
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
