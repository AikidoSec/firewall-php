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
	webscanner "main/vulnerabilities/web-scanner"
)

func OnPreRequest() string {
	context.Clear()
	return ""
}

func OnRequestShutdownReporting(server *ServerData, method, route, routeParsed string, statusCode int, user, username, userAgent, ip, rateLimitGroup string, apiSpec *protos.APISpec, rateLimited bool, queryParsed map[string]interface{}) {
	if method == "" || route == "" || statusCode == 0 {
		return
	}

	log.Info("[RSHUTDOWN] Got request metadata: ", method, " ", route, " ", statusCode)
	isWebScanner := webscanner.IsWebScanner(method, route, queryParsed)
	shouldDiscoverRoute := utils.ShouldDiscoverRoute(statusCode, route, method)
	if !rateLimited && !shouldDiscoverRoute && !isWebScanner {
		return
	}

	log.Info("[RSHUTDOWN] Got API spec: ", apiSpec)
	grpc.OnRequestShutdown(server, method, route, routeParsed, statusCode, user, username, userAgent, ip, rateLimitGroup, apiSpec, rateLimited, isWebScanner, shouldDiscoverRoute)
}

func OnPostRequest() string {
	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	go OnRequestShutdownReporting(server, context.GetMethod(), context.GetRoute(), context.GetParsedRoute(), context.GetStatusCode(), context.GetUserId(), context.GetUserName(), context.GetUserAgent(), context.GetIp(), context.GetRateLimitGroup(), api_discovery.GetApiInfo(server), context.IsEndpointRateLimited(), context.GetQueryParsed())
	context.Clear()
	return ""
}
