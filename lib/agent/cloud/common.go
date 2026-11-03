package cloud

import (
	"cmp"
	"encoding/json"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/log"
	"main/rate_limiting"
	"main/utils"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

func GetAgentInfo(server *ServerData) AgentInfo {
	return AgentInfo{
		DryMode:   !utils.IsBlockingEnabled(server),
		Hostname:  globals.Machine.HostName,
		Version:   constants.Version,
		IPAddress: globals.Machine.IPAddress,
		OS: OsInfo{
			Name:    globals.Machine.OS,
			Version: globals.Machine.OSVersion,
		},
		Platform: PlatformInfo{
			Name:    server.AikidoConfig.PlatformName,
			Version: server.AikidoConfig.PlatformVersion,
		},
		Packages: make(map[string]string, 0),
		NodeEnv:  "",
		Library:  "firewall-php",
	}
}

func setAgentHeaders(req *http.Request, server *ServerData) {
	req.Header.Set("X-Agent-Platform", "php")
	req.Header.Set("X-Agent-Library", "firewall-php")
	req.Header.Set("X-Agent-Version", constants.Version)
	req.Header.Set("X-Agent-Hostname", cmp.Or(globals.Machine.HostName, "unknown"))
	req.Header.Set("X-Agent-IP-Address", cmp.Or(globals.Machine.IPAddress, "unknown"))
	req.Header.Set("X-Agent-Session-Id", server.SessionID)
}

func ResetHeartbeatTicker(server *ServerData) {
	if !server.CloudConfig.ReceivedAnyStats {
		log.Info(server.Logger, "Resetting HeartbeatTicker to 1m!")
		server.PollingData.HeartbeatTicker.Reset(1 * time.Minute)
	} else {
		if server.CloudConfig.HeartbeatIntervalInMS >= constants.MinHeartbeatIntervalInMS {
			log.Infof(server.Logger, "Resetting HeartbeatTicker to %dms!", server.CloudConfig.HeartbeatIntervalInMS)
			server.PollingData.HeartbeatTicker.Reset(time.Duration(server.CloudConfig.HeartbeatIntervalInMS) * time.Millisecond)
		}
	}
}

func ApplyCloudConfig(server *ServerData) {
	log.Infof(server.Logger, "Applying new cloud config: %v", server.CloudConfig)
	ResetHeartbeatTicker(server)
	rate_limiting.UpdateConfig(server)
}

func UpdateIpsLists(ipLists []IpsData) map[string]IpBlocklist {
	m := make(map[string]IpBlocklist)
	for _, ipList := range ipLists {
		m[ipList.Key] = IpBlocklist{Description: ipList.Description, Ips: ipList.Ips}
	}
	return m
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

	server.CloudConfig.BlockedIpsList = UpdateIpsLists(tempListsConfig.BlockedIpAddresses)
	server.CloudConfig.MonitoredIpsList = UpdateIpsLists(tempListsConfig.MonitoredIpAddresses)
	server.CloudConfig.AllowedIpsList = UpdateIpsLists(tempListsConfig.AllowedIpAddresses)

	server.CloudConfig.BlockedUserAgents = tempListsConfig.BlockedUserAgents
	server.CloudConfig.MonitoredUserAgents = tempListsConfig.MonitoredUserAgents

	server.CloudConfig.UserAgentDetails = make(map[string]string)
	for _, userAgentDetail := range tempListsConfig.UserAgentDetails {
		server.CloudConfig.UserAgentDetails[userAgentDetail.Key] = userAgentDetail.Pattern
	}

	/* Force garbage collection to ensure that the IP blocklists temporary memory is released ASAP */
	tempListsConfig = ListsConfigData{}
	runtime.GC()

	return true
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

func LogCloudRequestError(server *ServerData, text string, err error) {
	if err.Error() == "no token set" && atomic.SwapUint32(&server.LoggedTokenError, 1) != 0 {
		return
	}
	log.Warn(server.Logger, text, err)
}
