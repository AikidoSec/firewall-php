package grpc

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"sort"
)

func storeStats() {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	globals.StatsData.Requests += 1
}

func storeAttackStats(req *protos.AttackDetected) {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	globals.StatsData.Attacks += 1
	if req.GetAttack().GetBlocked() {
		globals.StatsData.AttacksBlocked += 1
	}
}

func storeSinkStats(protoSinkStats *protos.MonitoredSinkStats) {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	sink := protoSinkStats.GetSink()
	monitoredSinkTimings, found := globals.StatsData.MonitoredSinkTimings[sink]
	if !found {
		monitoredSinkTimings = MonitoredSinkTimings{}
	}

	monitoredSinkTimings.AttacksDetected.Total += int(protoSinkStats.GetAttacksDetected())
	monitoredSinkTimings.AttacksDetected.Blocked += int(protoSinkStats.GetAttacksBlocked())
	monitoredSinkTimings.InterceptorThrewError += int(protoSinkStats.GetInterceptorThrewError())
	monitoredSinkTimings.WithoutContext += int(protoSinkStats.GetWithoutContext())
	monitoredSinkTimings.Total += int(protoSinkStats.GetTotal())
	monitoredSinkTimings.Timings = append(monitoredSinkTimings.Timings, protoSinkStats.GetTimings()...)

	globals.StatsData.MonitoredSinkTimings[sink] = monitoredSinkTimings
}

func getApiSpecData(apiSpec *protos.APISpec) (*protos.DataSchema, string, *protos.DataSchema, []*protos.APIAuthType) {
	if apiSpec == nil {
		return nil, "", nil, nil
	}

	var bodyDataSchema *protos.DataSchema = nil
	var bodyType string = ""
	if apiSpec.Body != nil {
		bodyDataSchema = apiSpec.Body.Schema
		bodyType = apiSpec.Body.Type
	}

	return bodyDataSchema, bodyType, apiSpec.Query, apiSpec.Auth
}

func getMergedApiSpec(currentApiSpec *protos.APISpec, newApiSpec *protos.APISpec) *protos.APISpec {
	if newApiSpec == nil {
		return currentApiSpec
	}
	if currentApiSpec == nil {
		return newApiSpec
	}

	currentBodySchema, currentBodyType, currentQuerySchema, currentAuth := getApiSpecData(currentApiSpec)
	newBodySchema, newBodyType, newQuerySchema, newAuth := getApiSpecData(newApiSpec)

	mergedBodySchema := api_discovery.MergeDataSchemas(currentBodySchema, newBodySchema)
	mergedQuerySchema := api_discovery.MergeDataSchemas(currentQuerySchema, newQuerySchema)
	mergedAuth := api_discovery.MergeApiAuthTypes(currentAuth, newAuth)
	if mergedBodySchema == nil && mergedQuerySchema == nil && mergedAuth == nil {
		return nil
	}

	mergedBodyType := newBodyType
	if mergedBodyType == "" {
		mergedBodyType = currentBodyType
	}

	return &protos.APISpec{
		Body: &protos.APIBodyInfo{
			Type:   mergedBodyType,
			Schema: mergedBodySchema,
		},
		Query: mergedQuerySchema,
		Auth:  mergedAuth,
	}
}

func storeRoute(method string, route string, apiSpec *protos.APISpec) {
	globals.RoutesMutex.Lock()
	defer globals.RoutesMutex.Unlock()

	if _, ok := globals.Routes[route]; !ok {
		globals.Routes[route] = make(map[string]*Route)
	}
	routeData, ok := globals.Routes[route][method]
	if !ok {
		routeData = &Route{Path: route, Method: method}
		globals.Routes[route][method] = routeData
	}

	routeData.Hits++
	routeData.ApiSpec = getMergedApiSpec(routeData.ApiSpec, apiSpec)
}

func incrementRateLimitingCounts(m map[string]*RateLimitingCounts, key string) {
	if key == "" {
		return
	}

	rateLimitingData, exists := m[key]
	if !exists {
		rateLimitingData = &RateLimitingCounts{}
		m[key] = rateLimitingData
	}

	rateLimitingData.TotalNumberOfRequests += 1
	rateLimitingData.NumberOfRequestsPerWindow.IncrementLast()
}

func updateRateLimitingCounts(method string, route string, user string, ip string) {
	globals.RateLimitingMutex.Lock()
	defer globals.RateLimitingMutex.Unlock()

	rateLimitingDataForEndpoint := getRateLimitingDataForEndpoint(method, route)
	if rateLimitingDataForEndpoint == nil {
		return
	}

	incrementRateLimitingCounts(rateLimitingDataForEndpoint.UserCounts, user)
	incrementRateLimitingCounts(rateLimitingDataForEndpoint.IpCounts, ip)
}

func isRateLimitingThresholdExceeded(config *RateLimitingConfig, countsMap map[string]*RateLimitingCounts, key string) bool {
	counts, exists := countsMap[key]
	if !exists {
		return false
	}

	return counts.TotalNumberOfRequests >= config.MaxRequests
}

func getRateLimitingValue(method, route string) *RateLimitingValue {
	rateLimitingDataForEndpoint, exists := globals.RateLimitingMap[RateLimitingKey{Method: method, Route: route}]
	if !exists {
		return nil
	}
	return rateLimitingDataForEndpoint
}

func getWildcardRateLimitingValues(method, route string) []*RateLimitingValue {
	wildcardRatelimitingValues := []*RateLimitingValue{}

	for key, r := range globals.RateLimitingWildcardMap {
		if key.Method != method {
			continue
		}
		if r.RouteRegex.MatchString(route) {
			wildcardRatelimitingValues = append(wildcardRatelimitingValues, r.RateLimitingValue)
		}
	}
	return wildcardRatelimitingValues
}

func getWildcardMatchingRateLimitingValues(method, route string) []*RateLimitingValue {
	rateLimitingDataArray := []*RateLimitingValue{}
	wildcardMethodRateLimitingData := getRateLimitingValue("*", route)
	if wildcardMethodRateLimitingData != nil {
		rateLimitingDataArray = append(rateLimitingDataArray, wildcardMethodRateLimitingData)
	}
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues(method, route)...)
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues("*", route)...)
	return rateLimitingDataArray
}

func getRateLimitingDataForEndpoint(method, route string) *RateLimitingValue {
	// Check for exact match first
	rateLimitingDataMatch := getRateLimitingValue(method, route)
	if rateLimitingDataMatch != nil {
		return rateLimitingDataMatch
	}

	// If no exact match, check for the most restrictive wildcard match

	wildcardMatches := getWildcardMatchingRateLimitingValues(method, route)
	if len(wildcardMatches) == 0 {
		return nil
	}

	sort.Slice(wildcardMatches, func(i, j int) bool {
		aRate := float64(wildcardMatches[i].Config.MaxRequests) / float64(wildcardMatches[i].Config.WindowSizeInMinutes)
		bRate := float64(wildcardMatches[j].Config.MaxRequests) / float64(wildcardMatches[j].Config.WindowSizeInMinutes)
		return aRate < bRate
	})

	return wildcardMatches[0]
}

func getRateLimitingStatus(method, route, user, ip string) *protos.RateLimitingStatus {
	globals.RateLimitingMutex.RLock()
	defer globals.RateLimitingMutex.RUnlock()

	rateLimitingDataMatch := getRateLimitingDataForEndpoint(method, route)

	if user != "" {
		// If the user exists, we only try to rate limit by user
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.UserCounts, user) {
			log.Infof("Rate limited request for user %s - %s %s - %v", user, method, route, rateLimitingDataMatch.UserCounts[user])
			return &protos.RateLimitingStatus{Block: true, Trigger: "user"}
		}
	} else {
		// Otherwise, we rate limit by ip
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.IpCounts, ip) {
			log.Infof("Rate limited request for ip %s - %s %s - %v", ip, method, route, rateLimitingDataMatch.IpCounts[ip])
			return &protos.RateLimitingStatus{Block: true, Trigger: "ip"}
		}
	}

	return &protos.RateLimitingStatus{Block: false}
}

func getCloudConfig(configUpdatedAt int64) *protos.CloudConfig {
	isBlockingEnabled := utils.IsBlockingEnabled()

	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if globals.CloudConfig.ConfigUpdatedAt <= configUpdatedAt {
		return nil
	}

	cloudConfig := &protos.CloudConfig{
		ConfigUpdatedAt:   globals.CloudConfig.ConfigUpdatedAt,
		BlockedUserIds:    globals.CloudConfig.BlockedUserIds,
		BypassedIps:       globals.CloudConfig.BypassedIps,
		BlockedIps:        map[string]*protos.IpList{},
		AllowedIps:        map[string]*protos.IpList{},
		BlockedUserAgents: globals.CloudConfig.BlockedUserAgents,
		Block:             isBlockingEnabled,
	}

	for ipBlocklistSource, ipBlocklist := range globals.CloudConfig.BlockedIpsList {
		cloudConfig.BlockedIps[ipBlocklistSource] = &protos.IpList{
			Description: ipBlocklist.Description,
			Ips:         ipBlocklist.Ips,
		}
	}

	for ipAllowlistSource, ipAllowlist := range globals.CloudConfig.AllowedIpsList {
		cloudConfig.AllowedIps[ipAllowlistSource] = &protos.IpList{
			Description: ipAllowlist.Description,
			Ips:         ipAllowlist.Ips,
		}
	}

	for _, endpoint := range globals.CloudConfig.Endpoints {
		cloudConfig.Endpoints = append(cloudConfig.Endpoints, &protos.Endpoint{
			Method:             endpoint.Method,
			Route:              endpoint.Route,
			ForceProtectionOff: endpoint.ForceProtectionOff,
			AllowedIPAddresses: endpoint.AllowedIPAddresses,
			RateLimiting: &protos.RateLimiting{
				Enabled: endpoint.RateLimiting.Enabled,
			},
		})
	}

	return cloudConfig
}

func onUserEvent(id string, username string, ip string) {
	globals.UsersMutex.Lock()
	defer globals.UsersMutex.Unlock()

	if _, exists := globals.Users[id]; exists {
		globals.Users[id] = User{
			ID:            id,
			Name:          username,
			LastIpAddress: ip,
			FirstSeenAt:   globals.Users[id].FirstSeenAt,
			LastSeenAt:    utils.GetTime(),
		}
		return
	}

	globals.Users[id] = User{
		ID:            id,
		Name:          username,
		LastIpAddress: ip,
		FirstSeenAt:   utils.GetTime(),
		LastSeenAt:    utils.GetTime(),
	}

}
