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
		"data":          data,
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

	if !globals.MiddlewareInstalled {
		go grpc.OnMiddlewareInstalled()
		globals.MiddlewareInstalled = true
	}

	userId := context.GetUserId()
	if utils.IsUserBlocked(userId) {
		log.Infof("User \"%s\" is blocked!", userId)
		return GetAction("store", "blocked", "user", "user blocked from config", userId, 403)
	}

	autoBlockingStatus := OnGetAutoBlockingStatus()

	if context.IsIpBypassed() {
		return ""
	}

	if context.IsEndpointRateLimitingEnabled() {
		// If request is monitored for rate limiting,
		// do a sync call via gRPC to see if the request should be blocked or not
		method := context.GetMethod()
		route := context.GetRoute()
		ip := context.GetIp()
		routeParsed := context.GetParsedRoute()
		if method == "" || route == "" {
			return ""
		}
		rateLimitingStatus := grpc.GetRateLimitingStatus(method, route, routeParsed, userId, ip, 10*time.Millisecond)
		if rateLimitingStatus != nil && rateLimitingStatus.Block {
			log.Infof("Request made from IP \"%s\" is ratelimited by \"%s\"!", ip, rateLimitingStatus.Trigger)
			return GetAction("store", "ratelimited", rateLimitingStatus.Trigger, "configured rate limit exceeded by current ip", ip, 429)
		}
	}

	return autoBlockingStatus
}

func OnGetAutoBlockingStatus() string {
	log.Debugf("OnGetAutoBlockingStatus called!")

	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	ip := context.GetIp()
	userAgent := context.GetUserAgent()

	if !context.IsEndpointIpAllowed() {
		log.Infof("IP \"%s\" is not allowd to access this endpoint!", ip)
		return GetAction("exit", "blocked", "ip", "not allowed by config to access this endpoint", ip, 403)
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
		return GetAction("exit", "blocked", "ip", ipBlockedDescriptions[0], ip, 403)
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
		return GetAction("exit", "blocked", "user-agent", description, userAgent, 403)
	}

	return ""
}
