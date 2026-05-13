package main

import (
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
)

func OnUserEvent(instance *instance.RequestProcessorInstance) string {
	id := context.GetUserId(instance)
	username := context.GetUserName(instance)
	ip := context.GetIp(instance)

	log.Infof(instance, "Got user event!")

	if id == "" || ip == "" {
		return ""
	}

	server := instance.GetCurrentServer()
	if server == nil {
		return ""
	}

	go grpc.OnUserEvent(server, instance.GetCurrentToken(), id, username, ip)
	return ""
}
