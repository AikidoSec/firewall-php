package main

import (
	"encoding/json"
	"fmt"
	"html"
	"main/context"
	"main/globals"
	"main/grpc"
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

func OnGetBlockingStatus() string {
	log.Debugf("OnGetBlockingStatus called!")

	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	if !server.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled(server)
		server.MiddlewareInstalled = true
	}

	userId := context.GetUserId()
	if utils.IsUserBlocked(server, userId) {
		log.Infof("User \"%s\" is blocked!", userId)
		return GetAction("store", "blocked", "user", "user blocked from config", userId, 403)
	}

	autoBlockingStatus := OnGetAutoBlockingStatus()

	if context.IsEndpointRateLimitingEnabled() {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		method := context.GetMethod()
		route := context.GetRoute()
		ip := context.GetIp()
		rateLimitGroup := context.GetRateLimitGroup()
		routeParsed := context.GetParsedRoute()
		if method == "" || route == "" {
			return ""
		}
		rateLimitingStatus := grpc.GetRateLimitingStatus(server, method, route, routeParsed, userId, ip, rateLimitGroup, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			context.ContextSetIsEndpointRateLimited()
			log.Infof("Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetAction("store", "ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip, 429)
		}
	}

	return autoBlockingStatus
}

func OnGetAutoBlockingStatus() string {
	log.Debugf("OnGetAutoBlockingStatus called!")

	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}
	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp()
	userAgent := context.GetUserAgent()

	if !context.IsEndpointIpAllowed() {
		log.Infof("IP \"%s\" is not allowed to access this endpoint!", ip)
		return GetAction("exit", "blocked", "ip", "not allowed by config to access this endpoint", ip, 403)
	}

	if !utils.IsIpAllowed(server, ip) {
		log.Infof("IP \"%s\" is not found in allow lists!", ip)
		return GetAction("exit", "blocked", "ip", "not in allow lists", ip, 403)
	}

	if ipMonitored, ipMonitoredMatches := utils.IsIpMonitored(server, ip); ipMonitored {
		log.Infof("IP \"%s\" found in monitored lists: %v!", ip, ipMonitoredMatches)
		go grpc.OnMonitoredIpMatch(server, ipMonitoredMatches)
	}

	if ipBlocked, ipBlockedMatches := utils.IsIpBlocked(server, ip); ipBlocked {
		log.Infof("IP \"%s\" found in blocked lists: %v!", ip, ipBlockedMatches)
		go grpc.OnMonitoredIpMatch(server, ipBlockedMatches)
		return GetAction("exit", "blocked", "ip", ipBlockedMatches[0].Description, ip, 403)
	}

	if userAgentMonitored, userAgentMonitoredDescriptions := utils.IsUserAgentMonitored(server, userAgent); userAgentMonitored {
		log.Infof("User Agent \"%s\" found in monitored lists: %v!", userAgent, userAgentMonitoredDescriptions)
		go grpc.OnMonitoredUserAgentMatch(server, userAgentMonitoredDescriptions)
	}

	if userAgentBlocked, userAgentBlockedDescriptions := utils.IsUserAgentBlocked(server, userAgent); userAgentBlocked {
		log.Infof("User Agent \"%s\" found in blocked lists: %v!", userAgent, userAgentBlockedDescriptions)
		go grpc.OnMonitoredUserAgentMatch(server, userAgentBlockedDescriptions)

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

func OnGetIsIpBypassed() string {
	log.Debugf("OnGetIsIpBypassed called!")
	if context.IsIpBypassed() {
		return GetBypassAction()
	}
	return ""
}
