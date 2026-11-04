package cloud

import (
	"cmp"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/log"
	"main/utils"
	"net/http"
	"sync/atomic"
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

func LogCloudRequestError(server *ServerData, text string, err error) {
	if err.Error() == "no token set" && atomic.SwapUint32(&server.LoggedTokenError, 1) != 0 {
		return
	}
	log.Warn(server.Logger, text, err)
}
