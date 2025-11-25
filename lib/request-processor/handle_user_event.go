package main

import (
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
)

func OnUserEvent(inst *instance.RequestProcessorInstance) string {
	id := context.GetUserId()
	username := context.GetUserName()
	ip := context.GetIp()

	log.Infof("Got user event!")

	if id == "" || ip == "" {
		return ""
	}

	server := inst.GetCurrentServer()
	if server == nil {
		return ""
	}
	go grpc.OnUserEvent(server, id, username, ip)
	return ""
}
