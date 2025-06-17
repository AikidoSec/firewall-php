package grpc

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"slices"
	"strings"
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

func storePackages(packages map[string]string) {
	globals.PackagesMutex.Lock()
	defer globals.PackagesMutex.Unlock()

	for packageName, packageVersion := range packages {
		globals.Packages[packageName] = Package{
			Name:       packageName,
			Version:    packageVersion,
			RequiredAt: utils.GetTime(),
		}
	}
}

func storeMonitoredListsMatches(m *map[string]int, lists []string) {
	if *m == nil {
		*m = make(map[string]int)
	}

	for _, list := range lists {
		if _, exists := (*m)[list]; !exists {
			(*m)[list] = 0
		}
		(*m)[list] += 1
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
		utils.RemoveOldestFromMapIfMaxExceeded(&globals.Routes, &globals.RoutesQueue, route)
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

func updateRateLimitingCounts(method string, route string, routeParsed string, user string, ip string) {
	globals.RateLimitingMutex.Lock()
	defer globals.RateLimitingMutex.Unlock()

	rateLimitingDataForEndpoint := getRateLimitingDataForEndpoint(method, route, routeParsed)
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

func getWildcardMatchingRateLimitingValues(method, route, routeParsed string) []*RateLimitingValue {
	rateLimitingDataArray := []*RateLimitingValue{}
	wildcardMethodRateLimitingData := getRateLimitingValue("*", routeParsed)
	if wildcardMethodRateLimitingData != nil {
		rateLimitingDataArray = append(rateLimitingDataArray, wildcardMethodRateLimitingData)
	}
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues(method, route)...)
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues("*", route)...)

	slices.SortFunc(rateLimitingDataArray, func(i, j *RateLimitingValue) int {
		// Sort endpoints based on the amount of * in the route
		return strings.Count(j.Route, "*") - strings.Count(i.Route, "*")
	})
	return rateLimitingDataArray
}

func getRateLimitingDataForEndpoint(method, route, routeParsed string) *RateLimitingValue {
	// Check for exact match first
	rateLimitingDataMatch := getRateLimitingValue(method, routeParsed)
	if rateLimitingDataMatch != nil {
		return rateLimitingDataMatch
	}

	// If no exact match, check for the most restrictive wildcard match
	wildcardMatches := getWildcardMatchingRateLimitingValues(method, route, routeParsed)
	if len(wildcardMatches) == 0 {
		return nil
	}

	slices.SortFunc(wildcardMatches, func(i, j *RateLimitingValue) int {
		aRate := float64(i.Config.MaxRequests) / float64(i.Config.WindowSizeInMinutes)
		bRate := float64(j.Config.MaxRequests) / float64(j.Config.WindowSizeInMinutes)
		return int(aRate - bRate)
	})

	return wildcardMatches[0]
}

func getRateLimitingStatus(method, route, routeParsed, user, ip string) *protos.RateLimitingStatus {
	globals.RateLimitingMutex.RLock()
	defer globals.RateLimitingMutex.RUnlock()

	rateLimitingDataMatch := getRateLimitingDataForEndpoint(method, route, routeParsed)
	if rateLimitingDataMatch == nil {
		return &protos.RateLimitingStatus{Block: false}
	}

	if user != "" {
		// If the user exists, we only try to rate limit by user
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.UserCounts, user) {
			log.Infof("Rate limited request for user %s - %s %s - %v", user, method, routeParsed, rateLimitingDataMatch.UserCounts[user])
			return &protos.RateLimitingStatus{Block: true, Trigger: "user"}
		}
	} else {
		// Otherwise, we rate limit by ip
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.IpCounts, ip) {
			log.Infof("Rate limited request for ip %s - %s %s - %v", ip, method, routeParsed, rateLimitingDataMatch.IpCounts[ip])
			return &protos.RateLimitingStatus{Block: true, Trigger: "ip"}
		}
	}

	return &protos.RateLimitingStatus{Block: false}
}

func getIpsList(ipsList map[string]IpBlocklist) map[string]*protos.IpList {
	m := make(map[string]*protos.IpList)
	for ipBlocklistSource, ipBlocklist := range ipsList {
		m[ipBlocklistSource] = &protos.IpList{Description: ipBlocklist.Description, Ips: ipBlocklist.Ips}
	}
	return m
}

func getCloudConfig(configUpdatedAt int64) *protos.CloudConfig {
	isBlockingEnabled := utils.IsBlockingEnabled()

	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if globals.CloudConfig.ConfigUpdatedAt <= configUpdatedAt {
		return nil
	}

	cloudConfig := &protos.CloudConfig{
		ConfigUpdatedAt:     globals.CloudConfig.ConfigUpdatedAt,
		BlockedUserIds:      globals.CloudConfig.BlockedUserIds,
		BypassedIps:         globals.CloudConfig.BypassedIps,
		BlockedIps:          getIpsList(globals.CloudConfig.BlockedIpsList),
		AllowedIps:          getIpsList(globals.CloudConfig.AllowedIpsList),
		BlockedUserAgents:   globals.CloudConfig.BlockedUserAgents,
		MonitoredIps:        getIpsList(globals.CloudConfig.MonitoredIpsList),
		MonitoredUserAgents: globals.CloudConfig.MonitoredUserAgents,
		UserAgentDetails:    globals.CloudConfig.UserAgentDetails,
		Block:               isBlockingEnabled,
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

	utils.RemoveOldestFromMapIfMaxExceeded(&globals.Users, &globals.UsersQueue, id)

	globals.Users[id] = User{
		ID:            id,
		Name:          username,
		LastIpAddress: ip,
		FirstSeenAt:   utils.GetTime(),
		LastSeenAt:    utils.GetTime(),
	}

}
