package main

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/context"
	"main/globals"
	"main/grpc"
	"main/ipc/protos"
	"main/log"
	"main/utils"
)

func OnPreRequest() string {
	context.Clear()
	return ""
}

func OnRequestShutdownReporting(server *ServerData, method, route, routeParsed string, statusCode int, user, ip, rateLimitGroup string, apiSpec *protos.APISpec, rateLimited bool) {
	if method == "" || route == "" || statusCode == 0 {
		return
	}

	log.Info("[RSHUTDOWN] Got request metadata: ", method, " ", route, " ", statusCode)

	if !rateLimited && !utils.ShouldDiscoverRoute(statusCode, route, method) {
		return
	}

	log.Info("[RSHUTDOWN] Got API spec: ", apiSpec)
	grpc.OnRequestShutdown(server, method, route, routeParsed, statusCode, user, ip, rateLimitGroup, apiSpec, rateLimited)
}

func OnPostRequest() string {
	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	go OnRequestShutdownReporting(server, context.GetMethod(), context.GetRoute(), context.GetParsedRoute(), context.GetStatusCode(), context.GetUserId(), context.GetIp(), context.GetRateLimitGroup(), api_discovery.GetApiInfo(server), context.IsEndpointRateLimited())
	context.Clear()
	return ""
}
