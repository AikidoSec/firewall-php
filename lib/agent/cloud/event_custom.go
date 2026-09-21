package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/ipc/protos"
)

func GetCustomEvent(server *ServerData, request *protos.CustomEvent) CustomEvent {
	event := CustomEvent{
		Type: "custom",
		Name: request.GetName(),
		Request: CustomEventRequest{
			Method:    request.GetRequest().GetMethod(),
			IPAddress: request.GetRequest().GetIpAddress(),
			UserAgent: request.GetRequest().GetUserAgent(),
			Source:    request.GetRequest().GetSource(),
			Route:     request.GetRequest().GetRoute(),
		},
		Agent: GetAgentInfo(server),
		Time:  request.GetTime(),
	}

	if request.GetUser() != nil {
		event.User = &CustomEventUser{
			ID:   request.GetUser().GetId(),
			Name: request.GetUser().GetName(),
		}
	}

	return event
}

func SendCustomEvent(server *ServerData, event CustomEvent) {
	// Do not log custom events because they contain data from normal requests.
	_, err := sendCloudRequest(server, server.AikidoConfig.Endpoint, constants.EventsAPI, constants.EventsAPIMethod, event, false)
	if err != nil {
		LogCloudRequestError(server, "Error sending custom event: ", err)
	}
}
