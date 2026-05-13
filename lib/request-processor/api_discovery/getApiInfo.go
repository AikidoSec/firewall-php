package api_discovery

import (
	. "main/aikido_types"
	"main/context"
	"main/instance"
	"main/ipc/protos"
	"main/log"
	"reflect"
)

func GetApiInfo(instance *instance.RequestProcessorInstance, server *ServerData) *protos.APISpec {
	if !server.AikidoConfig.CollectApiSchema {
		log.Debug(instance, "AIKIDO_FEATURE_COLLECT_API_SCHEMA is not enabled -> no API schema!")
		return nil
	}

	var bodyInfo *protos.APIBodyInfo
	var queryInfo *protos.DataSchema

	body := context.GetBodyParsed(instance)
	headers := context.GetHeadersParsed(instance)
	query := context.GetQueryParsed(instance)

	// Check body data
	if body != nil && isObject(body) && len(body) > 0 {
		bodyType := getBodyDataType(headers)
		if bodyType == Undefined {
			log.Debug(instance, "Body type is undefined -> no API schema!")
			return nil
		}

		bodySchema := GetDataSchema(body, 0)

		bodyInfo = &protos.APIBodyInfo{
			Type:   bodyType,
			Schema: bodySchema,
		}
	}

	// Check query data
	if query != nil && isObject(query) && len(query) > 0 {
		queryInfo = GetDataSchema(query, 0)
	}

	// Get Auth Info
	authInfo := GetApiAuthType(instance)

	if bodyInfo == nil && queryInfo == nil && authInfo == nil {
		log.Debug(instance, "All sub-schemas are empty -> no API schema!")
		return nil
	}

	return &protos.APISpec{
		Body:  bodyInfo,
		Query: queryInfo,
		Auth:  authInfo,
	}
}

func isObject(data interface{}) bool {
	if data == nil {
		return false
	}

	// Helper function to determine if the data is an object (map in Go)
	return reflect.TypeOf(data).Kind() == reflect.Map
}
