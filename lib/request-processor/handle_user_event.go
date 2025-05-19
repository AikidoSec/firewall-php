package main

import (
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
)

func OnUserEvent() string {
	if !globals.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled()
		globals.MiddlewareInstalled = true
	}

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
