package main

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
	"main/utils"
	webscanner "main/vulnerabilities/web-scanner"
)

func OnPreRequest(inst *instance.RequestProcessorInstance) string {
	context.Clear()
	return ""
}

func OnRequestShutdownReporting(params RequestShutdownParams) {
	if params.Method == "" || params.Route == "" || params.StatusCode == 0 {
		return
	}

	log.Info("[RSHUTDOWN] Got request metadata: ", params.Method, " ", params.Route, " ", params.StatusCode)
	params.IsWebScanner = webscanner.IsWebScanner(params.Method, params.Route, params.QueryParsed)
	params.ShouldDiscoverRoute = utils.ShouldDiscoverRoute(params.StatusCode, params.Route, params.Method)
	if !params.RateLimited && !params.ShouldDiscoverRoute && !params.IsWebScanner {
		return
	}

	log.Info("[RSHUTDOWN] Got API spec: ", params.APISpec)
	grpc.OnRequestShutdown(params)
}

func OnPostRequest(inst *instance.RequestProcessorInstance) string {
	server := inst.GetCurrentServer()
	if server == nil {
		return ""
	}

	params := RequestShutdownParams{
		Server:         server,
		Method:         context.GetMethod(),
		Route:          context.GetRoute(),
		RouteParsed:    context.GetParsedRoute(),
		StatusCode:     context.GetStatusCode(),
		User:           context.GetUserId(),
		UserAgent:      context.GetUserAgent(),
		IP:             context.GetIp(),
		RateLimitGroup: context.GetRateLimitGroup(),
		RateLimited:    context.IsEndpointRateLimited(),
		QueryParsed:    context.GetQueryParsed(),
		APISpec:        api_discovery.GetApiInfo(server), // Also needs context, must be called before Clear()
	}

	context.Clear()

	go func() {
		OnRequestShutdownReporting(params)
	}()

	return ""
}
