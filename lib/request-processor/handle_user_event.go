package main

import (
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
)

func OnUserEvent() string {
	id := context.GetUserId()
	username := context.GetUserName()
	ip := context.GetIp()

	log.Infof("Got user event!")

	if id == "" || ip == "" {
		return ""
	}

	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	go grpc.OnUserEvent(server, id, username, ip)
	return ""
}
