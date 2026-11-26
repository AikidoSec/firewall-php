package api_discovery

import (
	. "main/aikido_types"
	"main/context"
	"main/instance"
	"main/ipc/protos"
	"main/log"
	"reflect"
	"strconv"
	"strings"
)

const maxBodyBytesForSchema = 1024 * 1024 // 1 MiB

func GetApiInfo(instance *instance.RequestProcessorInstance, server *ServerData) *protos.APISpec {
	if !server.AikidoConfig.CollectApiSchema {
		log.Debug(instance, "AIKIDO_FEATURE_COLLECT_API_SCHEMA is not enabled -> no API schema!")
		return nil
	}

	headers := context.GetHeadersParsed(instance)
	contentLength, _ := headers["content_length"].(string)
	bodyBytes, err := strconv.ParseUint(strings.Trim(contentLength, " \t"), 10, 64)
	if bodyBytes > maxBodyBytesForSchema {
		return nil
	}

	var bodyInfo *protos.APIBodyInfo
	var queryInfo *protos.DataSchema
	var body map[string]interface{}
	if err == nil && bodyBytes > 0 {
		body = context.GetBodyParsed(instance)
	}
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
