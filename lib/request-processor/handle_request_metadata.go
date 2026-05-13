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

func OnPreRequest(instance *instance.RequestProcessorInstance) string {
	context.Clear(instance)
	return ""
}

func OnRequestShutdownReporting(params RequestShutdownParams) {
	if params.Method == "" || params.Route == "" || params.StatusCode == 0 {
		return
	}

	log.Info(nil, "[RSHUTDOWN] Got request metadata: ", params.Method, " ", params.Route, " ", params.StatusCode)

	params.IsWebScanner = webscanner.IsWebScanner(params.Method, params.Route, params.QueryParsed)

	params.ShouldDiscoverRoute = utils.ShouldDiscoverRoute(params.StatusCode, params.Route, params.Method)
	if !params.RateLimited && !params.ShouldDiscoverRoute && !params.IsWebScanner {
		return
	}

	log.Info(nil, "[RSHUTDOWN] Got API spec: ", params.APISpec)
	grpc.OnRequestShutdown(params)
}

func OnPostRequest(instance *instance.RequestProcessorInstance) string {
	server := instance.GetCurrentServer()
	if server == nil {
		return ""
	}

	// Only send request metadata if the IP is not bypassed
	if !context.IsIpBypassed(instance) {
		params := RequestShutdownParams{
			ThreadID:       instance.GetThreadID(),
			Token:          instance.GetCurrentToken(),
			Method:         context.GetMethod(instance),
			Route:          context.GetRoute(instance),
			RouteParsed:    context.GetParsedRoute(instance),
			StatusCode:     context.GetStatusCode(instance),
			User:           context.GetUserId(instance),
			UserAgent:      context.GetUserAgent(instance),
			IP:             context.GetIp(instance),
			Url:            context.GetUrl(instance),
			RateLimitGroup: context.GetRateLimitGroup(instance),
			RateLimited:    context.IsEndpointRateLimited(instance),
			QueryParsed:    context.GetQueryParsed(instance),
			APISpec:        api_discovery.GetApiInfo(instance, instance.GetCurrentServer()),
		}

		context.Clear(instance)

		go func() {
			OnRequestShutdownReporting(params)
		}()
	}

	return ""
}
