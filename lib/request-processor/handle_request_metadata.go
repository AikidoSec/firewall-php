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
	context.Clear(inst)
	return ""
}

func OnRequestShutdownReporting(params RequestShutdownParams) {
	if params.Method == "" || params.Route == "" || params.StatusCode == 0 {
		return
	}

	log.InfoWithThreadID(params.ThreadID, "[RSHUTDOWN] Got request metadata: ", params.Method, " ", params.Route, " ", params.StatusCode)
	// Only detect web scanner activity for non-bypassed IPs
	if !params.IsIpBypassed {
		params.IsWebScanner = webscanner.IsWebScanner(params.Method, params.Route, params.QueryParsed)
	}
	params.ShouldDiscoverRoute = utils.ShouldDiscoverRoute(params.StatusCode, params.Route, params.Method)
	if !params.RateLimited && !params.ShouldDiscoverRoute && !params.IsWebScanner {
		return
	}

	log.InfoWithThreadID(params.ThreadID, "[RSHUTDOWN] Got API spec: ", params.APISpec)
	grpc.OnRequestShutdown(params)
}

func OnPostRequest(inst *instance.RequestProcessorInstance) string {
	if inst.GetCurrentServer() == nil {
		return ""
	}
	if !context.IsIpBypassed(inst) {
		params := RequestShutdownParams{
			ThreadID:       inst.GetThreadID(),
			Token:          inst.GetCurrentToken(),
			Method:         context.GetMethod(inst),
			Route:          context.GetRoute(inst),
			RouteParsed:    context.GetParsedRoute(inst),
			StatusCode:     context.GetStatusCode(inst),
			User:           context.GetUserId(inst),
			UserAgent:      context.GetUserAgent(inst),
			IP:             context.GetIp(inst),
			Url:            context.GetUrl(inst),
			RateLimitGroup: context.GetRateLimitGroup(inst),
			RateLimited:    context.IsEndpointRateLimited(inst),
			QueryParsed:    context.GetQueryParsed(inst),
			IsIpBypassed:   context.IsIpBypassed(inst),
			APISpec:        api_discovery.GetApiInfo(inst, inst.GetCurrentServer()),
		}

		context.Clear(inst)

		go func() {
			OnRequestShutdownReporting(params)
		}()
	}

	return ""
}
