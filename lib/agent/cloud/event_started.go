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
		// This is the first cloud event, so GotTraffic can still be false. Log the
		// failure directly instead of suppressing the only startup error signal.
		log.Warn(server.Logger, "Error in sending start event: ", err)
		return
	}
	StoreCloudConfig(server, response)
}
