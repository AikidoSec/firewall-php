package cloud

import (
	. "main/aikido_types"
	"main/utils"
)

func SendStartEvent(server *ServerData) {
	startedEvent := Started{
		Type:  "started",
		Agent: GetAgentInfo(server),
		Time:  utils.GetTime(),
	}

	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, EventsAPI, EventsAPIMethod, startedEvent)
	if err != nil {
		LogCloudRequestError(server, "Error in sending start event: ", err)
		return
	}
	StoreCloudConfig(server, response)
}
