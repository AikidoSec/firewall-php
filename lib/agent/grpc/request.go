package grpc

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/constants"
	"main/ipc/protos"
	"main/log"
	"main/utils"
	"slices"
	"strings"
)

func storeTotalStats(server *ServerData, rateLimited bool) {
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	server.StatsData.Requests += 1
	if rateLimited {
		server.StatsData.RequestsRateLimited += 1
	}
}

func storeAttackStats(server *ServerData, req *protos.AttackDetected) {
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	server.StatsData.Attacks += 1
	if req.GetAttack().GetBlocked() {
		server.StatsData.AttacksBlocked += 1
	}
}

func storePackages(server *ServerData, packages map[string]string) {
	server.PackagesMutex.Lock()
	defer server.PackagesMutex.Unlock()

	for packageName, packageVersion := range packages {
		server.Packages[packageName] = Package{
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

func storeSinkStats(server *ServerData, protoSinkStats *protos.MonitoredSinkStats) {
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	sink := protoSinkStats.GetSink()
	monitoredSinkTimings, found := server.StatsData.MonitoredSinkTimings[sink]
	if !found {
		monitoredSinkTimings = MonitoredSinkTimings{}
	}

	monitoredSinkTimings.Kind = protoSinkStats.Kind
	monitoredSinkTimings.AttacksDetected.Total += int(protoSinkStats.GetAttacksDetected())
	monitoredSinkTimings.AttacksDetected.Blocked += int(protoSinkStats.GetAttacksBlocked())
	monitoredSinkTimings.InterceptorThrewError += int(protoSinkStats.GetInterceptorThrewError())
	monitoredSinkTimings.WithoutContext += int(protoSinkStats.GetWithoutContext())
	monitoredSinkTimings.Total += int(protoSinkStats.GetTotal())
	monitoredSinkTimings.Timings = append(monitoredSinkTimings.Timings, protoSinkStats.GetTimings()...)
	if len(monitoredSinkTimings.Timings) >= constants.MinStatsCollectedForRelevantMetrics {
		monitoredSinkTimings.CompressedTimings = append(monitoredSinkTimings.CompressedTimings, CompressedTiming{
			AverageInMS:  utils.ComputeAverage(monitoredSinkTimings.Timings),
			Percentiles:  utils.ComputePercentiles(monitoredSinkTimings.Timings),
			CompressedAt: utils.GetTime(),
		})
		monitoredSinkTimings.Timings = []int64{}
	}

	server.StatsData.MonitoredSinkTimings[sink] = monitoredSinkTimings
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

func storeRoute(server *ServerData, method string, route string, apiSpec *protos.APISpec, rateLimited bool) {
	server.RoutesMutex.Lock()
	defer server.RoutesMutex.Unlock()

	if _, ok := server.Routes[route]; !ok {
		server.Routes[route] = make(map[string]*Route)
		utils.RemoveOldestFromMapIfMaxExceeded(&server.Routes, &server.RoutesQueue, route)
	}

	routeData, ok := server.Routes[route][method]
	if !ok {
		routeData = &Route{Path: route, Method: method}
		server.Routes[route][method] = routeData
	}

	routeData.Hits++
	routeData.ApiSpec = getMergedApiSpec(routeData.ApiSpec, apiSpec)
	if rateLimited {
		routeData.RateLimitedCount += 1
	}
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

func updateRateLimitingCounts(server *ServerData, method string, route string, routeParsed string, user string, ip string, rateLimitGroup string) {
	server.RateLimitingMutex.Lock()
	defer server.RateLimitingMutex.Unlock()

	rateLimitingDataForEndpoint := getRateLimitingDataForEndpoint(server, method, route, routeParsed)
	if rateLimitingDataForEndpoint == nil {
		return
	}

	incrementRateLimitingCounts(rateLimitingDataForEndpoint.UserCounts, user)
	incrementRateLimitingCounts(rateLimitingDataForEndpoint.IpCounts, ip)
	incrementRateLimitingCounts(rateLimitingDataForEndpoint.RateLimitGroupCounts, rateLimitGroup)
}

func isRateLimitingThresholdExceeded(config *RateLimitingConfig, countsMap map[string]*RateLimitingCounts, key string) bool {
	counts, exists := countsMap[key]
	if !exists {
		return false
	}

	return counts.TotalNumberOfRequests >= config.MaxRequests
}

func getRateLimitingValue(server *ServerData, method, route string) *RateLimitingValue {
	rateLimitingDataForEndpoint, exists := server.RateLimitingMap[RateLimitingKey{Method: method, Route: route}]
	if !exists {
		return nil
	}
	return rateLimitingDataForEndpoint
}

func getWildcardRateLimitingValues(server *ServerData, method, route string) []*RateLimitingValue {
	wildcardRatelimitingValues := []*RateLimitingValue{}

	for key, r := range server.RateLimitingWildcardMap {
		if key.Method != method {
			continue
		}
		if r.RouteRegex.MatchString(route) {
			wildcardRatelimitingValues = append(wildcardRatelimitingValues, r.RateLimitingValue)
		}
	}
	return wildcardRatelimitingValues
}

func getWildcardMatchingRateLimitingValues(server *ServerData, method, route, routeParsed string) []*RateLimitingValue {
	rateLimitingDataArray := []*RateLimitingValue{}
	wildcardMethodRateLimitingData := getRateLimitingValue(server, "*", routeParsed)
	if wildcardMethodRateLimitingData != nil {
		rateLimitingDataArray = append(rateLimitingDataArray, wildcardMethodRateLimitingData)
	}
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues(server, method, route)...)
	rateLimitingDataArray = append(rateLimitingDataArray, getWildcardRateLimitingValues(server, "*", route)...)

	slices.SortFunc(rateLimitingDataArray, func(i, j *RateLimitingValue) int {
		// Sort endpoints based on the amount of * in the route
		return strings.Count(j.Route, "*") - strings.Count(i.Route, "*")
	})
	return rateLimitingDataArray
}

func getRateLimitingDataForEndpoint(server *ServerData, method, route, routeParsed string) *RateLimitingValue {
	// Check for exact match first
	rateLimitingDataMatch := getRateLimitingValue(server, method, routeParsed)
	if rateLimitingDataMatch != nil {
		return rateLimitingDataMatch
	}

	// If no exact match, check for the most restrictive wildcard match
	wildcardMatches := getWildcardMatchingRateLimitingValues(server, method, route, routeParsed)
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

func getRateLimitingStatus(server *ServerData, method, route, routeParsed, user, ip, rateLimitGroup string) *protos.RateLimitingStatus {
	server.RateLimitingMutex.RLock()
	defer server.RateLimitingMutex.RUnlock()

	rateLimitingDataMatch := getRateLimitingDataForEndpoint(server, method, route, routeParsed)
	if rateLimitingDataMatch == nil {
		return &protos.RateLimitingStatus{Block: false}
	}

	if rateLimitGroup != "" {
		// If the rate limit group exists, we only try to rate limit by rate limit group
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.RateLimitGroupCounts, rateLimitGroup) {
			log.Infof(server.Logger, "Rate limited request for group %s - %s %s - %v", rateLimitGroup, method, routeParsed, rateLimitingDataMatch.RateLimitGroupCounts[rateLimitGroup])
			return &protos.RateLimitingStatus{Block: true, Trigger: "group"}
		}
	} else if user != "" {
		// Otherwise, if the user exists, we try to rate limit by user
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.UserCounts, user) {
			log.Infof(server.Logger, "Rate limited request for user %s - %s %s - %v", user, method, routeParsed, rateLimitingDataMatch.UserCounts[user])
			return &protos.RateLimitingStatus{Block: true, Trigger: "user"}
		}
	} else {
		// Otherwise, we try to rate limit by ip
		if isRateLimitingThresholdExceeded(&rateLimitingDataMatch.Config, rateLimitingDataMatch.IpCounts, ip) {
			log.Infof(server.Logger, "Rate limited request for ip %s - %s %s - %v", ip, method, routeParsed, rateLimitingDataMatch.IpCounts[ip])
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

func getCloudConfig(server *ServerData, configUpdatedAt int64) *protos.CloudConfig {
	isBlockingEnabled := utils.IsBlockingEnabled(server)

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	if server.CloudConfig.ConfigUpdatedAt <= configUpdatedAt {
		return nil
	}

	cloudConfig := &protos.CloudConfig{
		ConfigUpdatedAt:     server.CloudConfig.ConfigUpdatedAt,
		BlockedUserIds:      server.CloudConfig.BlockedUserIds,
		BypassedIps:         server.CloudConfig.BypassedIps,
		BlockedIps:          getIpsList(server.CloudConfig.BlockedIpsList),
		AllowedIps:          getIpsList(server.CloudConfig.AllowedIpsList),
		BlockedUserAgents:   server.CloudConfig.BlockedUserAgents,
		MonitoredIps:        getIpsList(server.CloudConfig.MonitoredIpsList),
		MonitoredUserAgents: server.CloudConfig.MonitoredUserAgents,
		UserAgentDetails:    server.CloudConfig.UserAgentDetails,
		Block:               isBlockingEnabled,
	}

	for _, endpoint := range server.CloudConfig.Endpoints {
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

func onUserEvent(server *ServerData, id string, username string, ip string) {
	server.UsersMutex.Lock()
	defer server.UsersMutex.Unlock()

	if _, exists := server.Users[id]; exists {
		server.Users[id] = User{
			ID:            id,
			Name:          username,
			LastIpAddress: ip,
			FirstSeenAt:   server.Users[id].FirstSeenAt,
			LastSeenAt:    utils.GetTime(),
		}
		return
	}

	utils.RemoveOldestFromMapIfMaxExceeded(&server.Users, &server.UsersQueue, id)

	server.Users[id] = User{
		ID:            id,
		Name:          username,
		LastIpAddress: ip,
		FirstSeenAt:   utils.GetTime(),
		LastSeenAt:    utils.GetTime(),
	}

}
