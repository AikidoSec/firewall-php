package main

import (
	"encoding/json"
	"html"
	"main/context"
	"main/globals"
	"main/grpc"
	"main/log"
	"main/utils"
	"time"
)

func GetStoreAction(actionType, trigger, description, data string) string {
	actionMap := map[string]interface{}{
		"action":      "store",
		"type":        actionType,
		"trigger":     trigger,
		"description": html.EscapeString(description),
		trigger:       data,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func OnGetBlockingStatus() string {
	if !globals.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled()
		globals.MiddlewareInstalled = true
	}

	userId := context.GetUserId()
	if utils.IsUserBlocked(userId) {
		log.Infof("User \"%s\" is blocked!", userId)
		return GetStoreAction("blocked", "user", "user blocked from config", userId)
	}

	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp()
	userAgent := context.GetUserAgent()
	endpointData := utils.GetEndpointConfig(method, route)

	if endpointData != nil && !utils.IsIpAllowed(endpointData.AllowedIPAddresses, ip) {
		log.Infof("IP \"%s\" is not allowd to access this endpoint!", ip)
		return GetStoreAction("blocked", "ip", "not allowed by config to access this endpoint", ip)
	}

	if context.IsIpBypassed() {
		log.Infof("IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if ipMonitored, ipMonitoredDescriptions := utils.IsIpMonitored(ip); ipMonitored {
		log.Infof("IP \"%s\" found in monitored lists: %v!", ip, ipMonitoredDescriptions)
		go grpc.OnMonitoredIpMatch(ipMonitoredDescriptions)
	}

	if ipBlocked, ipBlockedDescriptions := utils.IsIpBlocked(ip); ipBlocked {
		log.Infof("IP \"%s\" found in blocked lists: %v!", ip, ipBlockedDescriptions)
		go grpc.OnMonitoredIpMatch(ipBlockedDescriptions)
		return GetStoreAction("blocked", "ip", ipBlockedDescriptions[0], ip)
	}

	if userAgentMonitored, userAgentMonitoredDescriptions := utils.IsUserAgentMonitored(userAgent); userAgentMonitored {
		log.Infof("User Agent \"%s\" found in monitored lists: %v!", userAgent, userAgentMonitoredDescriptions)
		go grpc.OnMonitoredUserAgentMatch(userAgentMonitoredDescriptions)
	}

	if userAgentBlocked, userAgentBlockedDescriptions := utils.IsUserAgentBlocked(userAgent); userAgentBlocked {
		log.Infof("User Agent \"%s\" found in blocked lists: %v!", userAgent, userAgentBlockedDescriptions)
		go grpc.OnMonitoredUserAgentMatch(userAgentBlockedDescriptions)

		description := "unknown"
		if len(userAgentBlockedDescriptions) > 0 {
			description = userAgentBlockedDescriptions[0]
		}
		return GetStoreAction("blocked", "user-agent", description, userAgent)
	}

	if endpointData != nil && endpointData.RateLimiting.Enabled {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		rateLimitingStatus := grpc.GetRateLimitingStatus(method, route, userId, ip, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			log.Infof("Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetStoreAction("ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip)
		}
	}

	return ""
}
