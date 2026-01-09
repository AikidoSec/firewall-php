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

	go grpc.OnUserEvent(inst, id, username, ip)
	return ""
}
