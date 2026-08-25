package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/log"
	"main/utils"
)

func SendStartEvent(server *ServerData) {
	startedEvent := Started{
		Type:  "started",
		Agent: GetAgentInfo(server),
		Time:  utils.GetTime(),
	}

	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.EventsAPI, constants.EventsAPIMethod, startedEvent)
	if err != nil {
		log.Warn(server.Logger, "Error in sending start event: ", err)
		return
	}
	StoreCloudConfig(server, response)
}
