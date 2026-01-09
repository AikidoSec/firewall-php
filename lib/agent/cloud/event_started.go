package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/utils"
	"sync/atomic"
)

func SendStartEvent(server *ServerData) {
	// In multi-worker mode (e.g., frankenphp-worker), ensure only one worker sends the started event
	// Use atomic compare-and-swap to guarantee exactly-once semantics
	if !atomic.CompareAndSwapUint32(&server.SentStartedEvent, 0, 1) {
		// Another worker already sent the started event
		return
	}

	startedEvent := Started{
		Type:  "started",
		Agent: GetAgentInfo(server),
		Time:  utils.GetTime(),
	}

	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, constants.EventsAPI, constants.EventsAPIMethod, startedEvent)
	if err != nil {
		LogCloudRequestError(server, "Error in sending start event: ", err)
		return
	}
	StoreCloudConfig(server, response)
}
