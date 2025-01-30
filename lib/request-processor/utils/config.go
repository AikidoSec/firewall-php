package utils

import (
	. "main/aikido_types"
	"main/globals"
)

func GetWildcardEndpointsConfigsForMethod(method string) []WildcardEndpointData {
	wildcardRoutesForMethod, found := globals.CloudConfig.WildcardEndpoints[method]
	if !found {
		return []WildcardEndpointData{}
	}
	return wildcardRoutesForMethod
}

func GetWildcardEndpointsConfigs(method string, route string) []EndpointData {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	wildcardRoutes := GetWildcardEndpointsConfigsForMethod(method)
	wildcardRoutes = append(wildcardRoutes, GetWildcardEndpointsConfigsForMethod("*")...)

	matchedEndpointsData := []EndpointData{}
	for _, wildcardEndpointData := range wildcardRoutes {
		if wildcardEndpointData.RouteRegex.MatchString(route) {
			matchedEndpointsData = append(matchedEndpointsData, wildcardEndpointData.Data)
		}
	}
	return matchedEndpointsData
}

func GetEndpointConfig(method string, route string) *EndpointData {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	endpointData, exists := globals.CloudConfig.Endpoints[EndpointKey{Method: method, Route: route}]
	if !exists {
		return nil
	}

	return &endpointData
}

func GetCloudConfigUpdatedAt() int64 {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	return globals.CloudConfig.ConfigUpdatedAt
}
