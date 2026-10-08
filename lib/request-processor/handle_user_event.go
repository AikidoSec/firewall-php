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

	userID := context.GetUserId(instance)
	userName := context.GetUserName(instance)
	ip := context.GetIp(instance)
	log.Infof(instance, "Got user event!")

	if userID == "" || ip == "" {
		return true
	}

	server := instance.GetCurrentServer()
	if server == nil {
		return true
	}

	go grpc.OnUserEvent(server, instance.GetCurrentToken(), userID, userName, ip)
	return true
}
