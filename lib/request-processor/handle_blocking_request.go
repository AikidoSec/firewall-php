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

	ip := context.GetIp()
	method := context.GetMethod()
	route := context.GetParsedRoute()
	if method == "" || route == "" {
		return ""
	}

	userId := context.GetUserId()

	userAgent := context.GetUserAgent()

	if ipAllowed, ipAllowedDescription := utils.IsIpAllowed(ip); !ipAllowed {
		// IP is NOT in the allowed IPs list
		log.Infof("IP \"%s\" is not allowed due to: %s!", ip, ipAllowedDescription)
		return GetStoreAction("blocked", "ip", ipAllowedDescription, ip)
	}

	if ipBlocked, ipBlockedDescription := utils.IsIpBlocked(ip); ipBlocked {
		// IP is in the blocked IPs list (TOR, bot, etc...)
		log.Infof("IP \"%s\" blocked due to: %s!", ip, ipBlockedDescription)
		return GetStoreAction("blocked", "ip", ipBlockedDescription, ip)
	}

	if userAgentBlocked, userAgentBlockedDescription := utils.IsUserAgentBlocked(userAgent); userAgentBlocked {
		// User Agent is in the blocked user agents list (known threat actors)
		log.Infof("User Agent \"%s\" blocked due to: %s!", userAgent, userAgentBlockedDescription)
		return GetStoreAction("blocked", "user-agent", userAgentBlockedDescription, userAgent)
	}

	if utils.IsUserBlocked(userId) {
		// User is blocked
		log.Infof("User \"%s\" is blocked!", userId)
		return GetStoreAction("blocked", "user", "user blocked from config", userId)
	}

	if context.IsIpBypassed() {
		// IP is bypassed
		log.Infof("IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if endpointData != nil && !utils.IsIpAllowedOnEndpoint(endpointData.AllowedIPAddresses, ip) {
		// IP is not allowed to access this endpoint
		log.Infof("IP \"%s\" is not allowed to access this endpoint!", ip)
		return GetStoreAction("blocked", "ip", "not allowed by config to access this endpoint", ip)
	}

	if userAgentBlocked, userAgentBlockedDescription := utils.IsUserAgentBlocked(userAgent); userAgentBlocked {
		// User Agent is in the blocked user agents list (known threat actors)
		log.Infof("User Agent \"%s\" blocked due to: %s!", userAgent, userAgentBlockedDescription)
		return GetStoreAction("blocked", "user-agent", userAgentBlockedDescription, userAgent)
	}

	if utils.IsUserBlocked(userId) {
		// User is blocked
		log.Infof("User \"%s\" is blocked!", userId)
		return GetStoreAction("blocked", "user", "user blocked from config", userId)
	}

	if context.IsIpBypassed() {
		// IP is bypassed
		log.Infof("IP \"%s\" is bypassed! Skipping additional checks...", ip)
		return ""
	}

	if endpointData != nil && !utils.IsIpAllowedOnEndpoint(endpointData.AllowedIPAddresses, ip) {
		// IP is not allowed to access this endpoint
		log.Infof("IP \"%s\" is not allowed to access this endpoint!", ip)
		return GetStoreAction("blocked", "ip", "not allowed by config to access this endpoint", ip)
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
