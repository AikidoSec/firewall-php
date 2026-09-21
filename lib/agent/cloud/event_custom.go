package cloud

import (
	. "main/aikido_types"
	"main/constants"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"sync/atomic"
)

// Shared by all servers, so a slow cloud cannot pile up goroutines and sockets in the agent.
var customEventSlots = make(chan struct{}, constants.MaxConcurrentCustomEventRequests)

var lastCustomEventDropWarningAt atomic.Int64

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

func ScheduleCustomEvent(server *ServerData, event CustomEvent) {
	select {
	case customEventSlots <- struct{}{}:
	default:
		warnCustomEventDropped(server)
		return
	}

	go func() {
		defer func() { <-customEventSlots }()
		sendCustomEvent(server, event)
	}()
}

func warnCustomEventDropped(server *ServerData) {
	now := utils.GetTime()
	last := lastCustomEventDropWarningAt.Load()
	if now-last < constants.CustomEventDropWarningIntervalInMs || !lastCustomEventDropWarningAt.CompareAndSwap(last, now) {
		return
	}
	log.Warnf(server.Logger, "Dropped custom event because %d custom events are already being sent. Playbooks might not trigger.", constants.MaxConcurrentCustomEventRequests)
}

func sendCustomEvent(server *ServerData, event CustomEvent) {
	// Do not log the payload. It can contain private data from normal requests.
	_, err := sendCloudRequest(server, server.AikidoConfig.Endpoint, constants.EventsAPI, constants.EventsAPIMethod, event, false)
	if err != nil {
		LogCloudRequestError(server, "Error sending custom event: ", err)
	}
}
