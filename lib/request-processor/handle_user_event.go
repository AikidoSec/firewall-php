package main

import (
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
)

func OnUserEvent(inst *instance.RequestProcessorInstance) string {
	id := context.GetUserId(inst)
	username := context.GetUserName(inst)
	ip := context.GetIp(inst)

	log.Infof(inst, "Got user event!")

	if id == "" || ip == "" {
		return ""
	}

	server := inst.GetCurrentServer()
	if server == nil {
		return ""
	}

	go grpc.OnUserEvent(inst.GetThreadID(), server, inst.GetCurrentToken(), id, username, ip)
	return ""
}
