package main

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/utils"
	webscanner "main/vulnerabilities/web-scanner"
)

func OnPreRequest() string {
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

func OnPostRequest() string {
	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	go OnRequestShutdownReporting(RequestShutdownParams{
		Server:         server,
		Method:         context.GetMethod(),
		Route:          context.GetRoute(),
		RouteParsed:    context.GetParsedRoute(),
		StatusCode:     context.GetStatusCode(),
		User:           context.GetUserId(),
		UserAgent:      context.GetUserAgent(),
		IP:             context.GetIp(),
		Url:            context.GetUrl(),
		RateLimitGroup: context.GetRateLimitGroup(),
		APISpec:        api_discovery.GetApiInfo(server),
		RateLimited:    context.IsEndpointRateLimited(),
		QueryParsed:    context.GetQueryParsed(),
	})
	context.Clear()
	return ""
}
