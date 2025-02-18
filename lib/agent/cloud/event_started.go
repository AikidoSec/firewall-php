package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/utils"
)

func SendStartEvent() {
	startedEvent := Started{
		Type:  "started",
		Agent: GetAgentInfo(),
		Time:  utils.GetTime(),
	}

	response, err := SendCloudRequest(globals.EnvironmentConfig.Endpoint, globals.EventsAPI, globals.EventsAPIMethod, startedEvent)
	if err != nil {
		LogCloudRequestError("Error in sending start event: ", err)
		return
	}
	StoreCloudConfig(response)
}
