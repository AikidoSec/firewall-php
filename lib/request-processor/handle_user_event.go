package main

import (
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
)

func OnUserEvent(instance *instance.RequestProcessorInstance, id string, username string) bool {
	if !context.SetUser(instance, id, username) {
		return false
	}

	id = context.GetUserId(instance)
	username = context.GetUserName(instance)
	ip := context.GetIp(instance)
	log.Infof(instance, "Got user event!")

	if id == "" || ip == "" {
		return true
	}

	server := instance.GetCurrentServer()
	if server == nil {
		return true
	}

	go grpc.OnUserEvent(server, instance.GetCurrentToken(), id, username, ip)
	return true
}
