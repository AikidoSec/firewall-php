package main

import (
	"encoding/json"
	"fmt"
	"html"
	"main/context"
	"main/grpc"
	"main/instance"
	"main/log"
	"main/utils"
	"time"
)

func GetAction(actionHandling, actionType, trigger, description, data string, responseCode int) string {
	actionMap := map[string]interface{}{
		"action":        actionHandling,
		"type":          actionType,
		"trigger":       trigger,
		"description":   html.EscapeString(description),
		"message":       fmt.Sprintf("Your %s (%s) is blocked due to: %s!", trigger, data, description),
		trigger:         data,
		"response_code": responseCode,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func OnGetBlockingStatus(inst *instance.RequestProcessorInstance) string {
	log.Debugf(inst, "OnGetBlockingStatus called!")

	server := inst.GetCurrentServer()
	if server == nil {
		return ""
	}
	if !server.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled(inst.GetThreadID(), server, inst.GetCurrentToken())
		server.MiddlewareInstalled = true
	}

	userId := context.GetUserId(inst)
	if utils.IsUserBlocked(server, userId) {
		log.Infof(inst, "User \"%s\" is blocked!", userId)
		return GetAction("store", "blocked", "user", "user blocked from config", userId, 403)
	}

	autoBlockingStatus := OnGetAutoBlockingStatus(inst)

	if context.IsIpBypassed(inst) {
		return ""
	}

	if context.IsEndpointRateLimitingEnabled(inst) {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		method := context.GetMethod(inst)
		route := context.GetRoute(inst)
		ip := context.GetIp(inst)
		rateLimitGroup := context.GetRateLimitGroup(inst)
		routeParsed := context.GetParsedRoute(inst)
		if method == "" || route == "" {
			return ""
		}
		rateLimitingStatus := grpc.GetRateLimitingStatus(inst, server, method, route, routeParsed, userId, ip, rateLimitGroup, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			context.ContextSetIsEndpointRateLimited(inst)
			log.Infof(inst, "Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetAction("store", "ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip, 429)
		}
	}

	return autoBlockingStatus
}

func OnGetAutoBlockingStatus(inst *instance.RequestProcessorInstance) string {
	log.Debugf(inst, "OnGetAutoBlockingStatus called!")

	server := inst.GetCurrentServer()
	if server == nil {
		return ""
	}
	method := context.GetMethod(inst)
	route := context.GetParsedRoute(inst)
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp(inst)
	userAgent := context.GetUserAgent(inst)

	if !context.IsEndpointIpAllowed(inst) {
		log.Infof(inst, "IP \"%s\" is not allowed to access this endpoint!", ip)
		return GetAction("exit", "blocked", "ip", "not allowed by config to access this endpoint", ip, 403)
	}

	if context.IsIpBypassed(inst) {
		log.Infof(inst, "IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if !utils.IsIpAllowed(inst, server, ip) {
		log.Infof(inst, "IP \"%s\" is not found in allow lists!", ip)
		return GetAction("exit", "blocked", "ip", "not in allow lists", ip, 403)
	}

	if ipMonitored, ipMonitoredMatches := utils.IsIpMonitored(inst, server, ip); ipMonitored {
		log.Infof(inst, "IP \"%s\" found in monitored lists: %v!", ip, ipMonitoredMatches)
		go grpc.OnMonitoredIpMatch(inst.GetThreadID(), server, inst.GetCurrentToken(), ipMonitoredMatches)
	}

	if ipBlocked, ipBlockedMatches := utils.IsIpBlocked(inst, server, ip); ipBlocked {
		log.Infof(inst, "IP \"%s\" found in blocked lists: %v!", ip, ipBlockedMatches)
		go grpc.OnMonitoredIpMatch(inst.GetThreadID(), server, inst.GetCurrentToken(), ipBlockedMatches)
		return GetAction("exit", "blocked", "ip", ipBlockedMatches[0].Description, ip, 403)
	}

	if userAgentMonitored, userAgentMonitoredDescriptions := utils.IsUserAgentMonitored(server, userAgent); userAgentMonitored {
		log.Infof(inst, "User Agent \"%s\" found in monitored lists: %v!", userAgent, userAgentMonitoredDescriptions)
		go grpc.OnMonitoredUserAgentMatch(inst.GetThreadID(), server, inst.GetCurrentToken(), userAgentMonitoredDescriptions)
	}

	if userAgentBlocked, userAgentBlockedDescriptions := utils.IsUserAgentBlocked(server, userAgent); userAgentBlocked {
		log.Infof(inst, "User Agent \"%s\" found in blocked lists: %v!", userAgent, userAgentBlockedDescriptions)
		go grpc.OnMonitoredUserAgentMatch(inst.GetThreadID(), server, inst.GetCurrentToken(), userAgentBlockedDescriptions)

		description := "unknown"
		if len(userAgentBlockedDescriptions) > 0 {
			description = userAgentBlockedDescriptions[0]
		}
		return GetAction("exit", "blocked", "user-agent", description, userAgent, 403)
	}

	return ""
}

func GetBypassAction() string {
	actionMap := map[string]interface{}{
		"action": "bypassIp",
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func OnGetIsIpBypassed(inst *instance.RequestProcessorInstance) string {
	log.Debugf(inst, "OnGetIsIpBypassed called!")
	if context.IsIpBypassed(inst) {
		return GetBypassAction()
	}
	return ""
}
