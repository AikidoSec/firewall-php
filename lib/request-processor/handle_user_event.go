package main

import (
	"main/grpc"
	"main/log"
	"main/utils"
)

func OnUserEvent(data map[string]interface{}) string {
	id := utils.MustGetFromMap[string](data, "id")
	username := utils.MustGetFromMap[string](data, "username")
	remoteAddress := utils.MustGetFromMap[string](data, "remoteAddress")
	xForwardedFor := utils.MustGetFromMap[string](data, "xForwardedFor")

	log.Infof("[UEVENT] Got user event: %s %s %s %s", id, username, remoteAddress, xForwardedFor)

	ip := utils.GetIpFromRequest(remoteAddress, xForwardedFor)

	log.Infof("[UEVENT] Got IP from request: %s", ip)

	if id == "" || username == "" || ip == "" {
		return "{\"status\": \"ok\"}"
	}

	go grpc.OnUserEvent(id, username, ip)

	return "{\"status\": \"ok\"}"
}
