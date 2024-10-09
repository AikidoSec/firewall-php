package grpc

import (
	. "main/aikido_types"
	"main/api_discovery"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
)

func storeStats() {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	globals.StatsData.Requests += 1
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

func updateRateLimitingStatus(method string, route string) {
	globals.RateLimitingMutex.Lock()
	defer globals.RateLimitingMutex.Unlock()

	rateLimitingData, exists := globals.RateLimitingMap[RateLimitingKey{Method: method, Route: route}]
	if !exists {
		return
	}

	rateLimitingData.Status.TotalNumberOfRequests += 1
	rateLimitingData.Status.NumberOfRequestsPerWindow.IncrementLast()
}

func getRequestStatus(method string, route string) *protos.RequestStatus {
	globals.RateLimitingMutex.Lock()
	defer globals.RateLimitingMutex.Unlock()

	forwardToServer := true

	rateLimitingData, exists := globals.RateLimitingMap[RateLimitingKey{Method: method, Route: route}]
	if exists && rateLimitingData.Status.TotalNumberOfRequests >= rateLimitingData.Config.MaxRequests {
		log.Infof("Rate limited request for (%s, %s) - status (%v)", method, route, rateLimitingData)
		forwardToServer = false
	}

	return &protos.RequestStatus{
		ForwardToServer: forwardToServer,
	}
}

func getCloudConfig() *protos.CloudConfig {
	isBlockingEnabled := utils.IsBlockingEnabled()

	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	cloudConfig := &protos.CloudConfig{
		BlockedUserIds: globals.CloudConfig.BlockedUserIds,
		BypassedIps:    globals.CloudConfig.BypassedIps,
		Block:          isBlockingEnabled,
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
