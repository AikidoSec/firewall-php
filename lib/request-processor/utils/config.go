package utils

import (
	. "main/aikido_types"
)

func GetWildcardEndpointsConfigsForMethod(server *ServerData, method string) []WildcardEndpointData {
	wildcardRoutesForMethod, found := server.CloudConfig.WildcardEndpoints[method]
	if !found {
		return []WildcardEndpointData{}
	}
	return wildcardRoutesForMethod
}

func GetWildcardEndpointsConfigs(server *ServerData, method string, route string) []EndpointData {
	if server == nil {
		return []EndpointData{}
	}
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	// We prioritize defined methods over wildcard methods
	wildcardRoutes := GetWildcardEndpointsConfigsForMethod(server, method)
	wildcardRoutes = append(wildcardRoutes, GetWildcardEndpointsConfigsForMethod(server, "*")...)

	matchedEndpointsData := []EndpointData{}
	for _, wildcardEndpointData := range wildcardRoutes {
		if wildcardEndpointData.RouteRegex.MatchString(route) {
			matchedEndpointsData = append(matchedEndpointsData, wildcardEndpointData.Data)
		}
	}
	return matchedEndpointsData
}

func GetEndpointConfig(server *ServerData, method string, route string) *EndpointData {
	if server == nil {
		return nil
	}
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	endpointData, exists := server.CloudConfig.Endpoints[EndpointKey{Method: method, Route: route}]
	if !exists {
		return nil
	}

	return &endpointData
}

func GetCloudConfigUpdatedAt(server *ServerData) int64 {
	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	return server.CloudConfig.ConfigUpdatedAt
}
