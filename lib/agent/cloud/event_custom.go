package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/ipc/protos"
	"main/log"
	"main/utils"
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

func ScheduleCustomEvent(server *ServerData, request *protos.CustomEvent) {
	if server.CustomEventsInFlight.Add(1) > constants.MaxConcurrentCustomEventRequests {
		server.CustomEventsInFlight.Add(-1)
		warnCustomEventDropped(server)
		return
	}

	go func() {
		defer server.CustomEventsInFlight.Add(-1)
		sendCustomEvent(server, GetCustomEvent(server, request))
	}()
}

func warnCustomEventDropped(server *ServerData) {
	now := utils.GetTime()
	last := server.LastCustomEventDropWarningAt.Load()
	if now-last < constants.CustomEventDropWarningIntervalInMs || !server.LastCustomEventDropWarningAt.CompareAndSwap(last, now) {
		return
	}
	log.Warnf(server.Logger, "Dropped custom event because %d custom events are already being processed for this server. Playbooks might not trigger.", constants.MaxConcurrentCustomEventRequests)
}

func sendCustomEvent(server *ServerData, event CustomEvent) {
	// Do not log the payload. It can contain private data from normal requests.
	_, err := sendCloudRequest(server, server.AikidoConfig.Endpoint, constants.EventsAPI, constants.EventsAPIMethod, event, false)
	if err != nil {
		LogCloudRequestError(server, "Error sending custom event: ", err)
	}
}
