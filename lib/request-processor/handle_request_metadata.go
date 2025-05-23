package main

import (
	"main/api_discovery"
	"main/context"
	"main/grpc"
	"main/ipc/protos"
	"main/log"
	"main/utils"
)

func OnPreRequest() string {
	context.Clear()
	return ""
}

func OnRequestShutdownReporting(method string, route string, routeParsed string, statusCode int, user string, ip string, apiSpec *protos.APISpec) {
	if method == "" || route == "" || statusCode == 0 {
		return
	}

	log.Info("[RSHUTDOWN] Got request metadata: ", method, " ", route, " ", statusCode)

	if !utils.ShouldDiscoverRoute(statusCode, route, method) {
		return
	}

	log.Info("[RSHUTDOWN] Got API spec: ", apiSpec)
	grpc.OnRequestShutdown(method, route, routeParsed, statusCode, user, ip, apiSpec)
}

func OnPostRequest() string {
	go OnRequestShutdownReporting(context.GetMethod(), context.GetRoute(), context.GetParsedRoute(), context.GetStatusCode(), context.GetUserId(), context.GetIp(), api_discovery.GetApiInfo())
	context.Clear()
	return ""
}
