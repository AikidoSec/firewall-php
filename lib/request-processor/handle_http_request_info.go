package main

import (
	"main/grpc"
	"main/log"
	"main/utils"
)

func OnRequestMetadata(data map[string]interface{}) string {
	method := utils.MustGetFromMap[string](data, "method")
	route := utils.MustGetFromMap[string](data, "route")

	if method == "" || route == "" {
		log.Error("Missing required fields in OnRequestMetadata")
		return "{\"status\": \"ok\"}"
	}

	log.Info("Got HTTP request: ", method, " ", route)
	go grpc.OnReceiveRequestMetadata(method, route)

	return "{\"status\": \"ok\"}"
}
