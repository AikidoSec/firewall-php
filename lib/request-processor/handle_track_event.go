package main

import (
	"main/context"
	"main/grpc"
	"main/instance"
	"main/ipc/protos"
	"time"
)

func OnTrackEvent(instance *instance.RequestProcessorInstance) string {
	name := context.GetCustomEventName(instance)
	server := instance.GetCurrentServer()
	if name == "" || server == nil {
		return ""
	}

	event := &protos.CustomEvent{
		Token:     instance.GetCurrentToken(),
		ServerPid: context.GetServerPID(),
		Name:      name,
		Request: &protos.Request{
			Method:    context.GetMethod(instance),
			IpAddress: context.GetIp(instance),
			UserAgent: context.GetUserAgent(instance),
			Source:    "php",
			Route:     context.GetRoute(instance),
		},
		Time: time.Now().UnixMilli(),
	}

	if userID := context.GetUserId(instance); userID != "" {
		event.User = &protos.CustomEventUser{
			Id:   userID,
			Name: context.GetUserName(instance),
		}
	}

	go grpc.OnCustomEvent(server, event)
	return ""
}
