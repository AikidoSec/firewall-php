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

func OnGetBlockingStatus(instance *instance.RequestProcessorInstance) string {
	log.Debugf(instance, "OnGetBlockingStatus called!")

	server := instance.GetCurrentServer()
	if server == nil {
		return ""
	}
	if !server.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled(server, instance.GetCurrentToken())
		server.MiddlewareInstalled = true
	}

	userId := context.GetUserId(instance)
	if utils.IsUserBlocked(server, userId) {
		log.Infof(instance, "User \"%s\" is blocked!", userId)
		return GetAction("store", "blocked", "user", "user blocked from config", userId, 403)
	}

	autoBlockingStatus := OnGetAutoBlockingStatus(instance)

	if context.IsIpBypassed(instance) {
		return ""
	}

	if context.IsEndpointRateLimitingEnabled(instance) {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		method := context.GetMethod(instance)
		route := context.GetRoute(instance)
		ip := context.GetIp(instance)
		rateLimitGroup := context.GetRateLimitGroup(instance)
		routeParsed := context.GetParsedRoute(instance)
		if method == "" || route == "" {
			return ""
		}
		rateLimitingStatus := grpc.GetRateLimitingStatus(instance, server, method, route, routeParsed, userId, ip, rateLimitGroup, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			context.ContextSetIsEndpointRateLimited(instance)
			log.Infof(instance, "Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetAction("store", "ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip, 429)
		}
	}

	return autoBlockingStatus
}

func OnGetAutoBlockingStatus(instance *instance.RequestProcessorInstance) string {
	log.Debugf(instance, "OnGetAutoBlockingStatus called!")

	server := instance.GetCurrentServer()
	if server == nil {
		return ""
	}
	method := context.GetMethod(instance)
	route := context.GetParsedRoute(instance)
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp(instance)
	userAgent := context.GetUserAgent(instance)

	if !context.IsEndpointIpAllowed(instance) {
		log.Infof(instance, "IP \"%s\" is not allowed to access this endpoint!", ip)
		return GetAction("exit", "blocked", "ip", "not allowed by config to access this endpoint", ip, 403)
	}

	if context.IsIpBypassed(instance) {
		log.Infof(instance, "IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if !utils.IsIpAllowed(instance, server, ip) {
		log.Infof(instance, "IP \"%s\" is not found in allow lists!", ip)
		return GetAction("exit", "blocked", "ip", "not in allow lists", ip, 403)
	}

	if ipMonitored, ipMonitoredMatches := utils.IsIpMonitored(instance, server, ip); ipMonitored {
		log.Infof(instance, "IP \"%s\" found in monitored lists: %v!", ip, ipMonitoredMatches)
		go grpc.OnMonitoredIpMatch(server, instance.GetCurrentToken(), ipMonitoredMatches)
	}

	if ipBlocked, ipBlockedMatches := utils.IsIpBlocked(instance, server, ip); ipBlocked {
		log.Infof(instance, "IP \"%s\" found in blocked lists: %v!", ip, ipBlockedMatches)
		go grpc.OnMonitoredIpMatch(server, instance.GetCurrentToken(), ipBlockedMatches)
		return GetAction("exit", "blocked", "ip", ipBlockedMatches[0].Description, ip, 403)
	}

	if userAgentMonitored, userAgentMonitoredDescriptions := utils.IsUserAgentMonitored(server, userAgent); userAgentMonitored {
		log.Infof(instance, "User Agent \"%s\" found in monitored lists: %v!", userAgent, userAgentMonitoredDescriptions)
		go grpc.OnMonitoredUserAgentMatch(server, instance.GetCurrentToken(), userAgentMonitoredDescriptions)
	}

	if userAgentBlocked, userAgentBlockedDescriptions := utils.IsUserAgentBlocked(server, userAgent); userAgentBlocked {
		log.Infof(instance, "User Agent \"%s\" found in blocked lists: %v!", userAgent, userAgentBlockedDescriptions)
		go grpc.OnMonitoredUserAgentMatch(server, instance.GetCurrentToken(), userAgentBlockedDescriptions)

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

func OnGetIsIpBypassed(instance *instance.RequestProcessorInstance) string {
	log.Debugf(instance, "OnGetIsIpBypassed called!")
	if context.IsIpBypassed(instance) {
		return GetBypassAction()
	}
	return ""
}
