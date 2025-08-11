package main

import (
	"main/context"
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

	go grpc.OnUserEvent(id, username, ip)
	return ""
}
