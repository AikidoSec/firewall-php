package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/utils"
)

func GetHeaders(protoHeaders []*protos.Header) map[string][]string {
	headers := map[string][]string{}

	for _, header := range protoHeaders {
		headers[header.Key] = []string{header.Value}
	}
	return headers
}

func GetMetadata(protoMetadata []*protos.Metadata) map[string]string {
	metas := map[string]string{}

	for _, meta := range protoMetadata {
		metas[meta.Key] = meta.Value
	}
	return metas
}

func GetRequestInfo(protoRequest *protos.Request) RequestInfo {
	return RequestInfo{
		Method:    protoRequest.Method,
		IPAddress: protoRequest.IpAddress,
		UserAgent: protoRequest.UserAgent,
		URL:       protoRequest.Url,
		Headers:   GetHeaders(protoRequest.Headers),
		Body:      protoRequest.Body,
		Source:    protoRequest.Source,
		Route:     protoRequest.Route,
	}
}

func GetAttackDetails(server *ServerData, protoAttack *protos.Attack) AttackDetails {
	return AttackDetails{
		Kind:      protoAttack.Kind,
		Operation: protoAttack.Operation,
		Module:    protoAttack.Module,
		Blocked:   protoAttack.Blocked,
		Source:    protoAttack.Source,
		Path:      protoAttack.Path,
		Stack:     protoAttack.Stack,
		Payload:   protoAttack.Payload,
		Metadata:  GetMetadata(protoAttack.Metadata),
		User:      utils.GetUserById(server, protoAttack.UserId),
	}
}

func ShouldSendAttackDetectedEvent(server *ServerData) bool {
	server.AttackDetectedEventsSentAtMutex.Lock()
	defer server.AttackDetectedEventsSentAtMutex.Unlock()

	currentTime := utils.GetTime()

	// Filter out events that are outside the current interval
	var filteredEvents []int64
	for _, eventTime := range server.AttackDetectedEventsSentAt {
		if eventTime > currentTime-globals.AttackDetectedEventsIntervalInMs {
			filteredEvents = append(filteredEvents, eventTime)
		}
	}
	server.AttackDetectedEventsSentAt = filteredEvents

	if len(server.AttackDetectedEventsSentAt) >= globals.MaxAttackDetectedEventsPerInterval {
		log.Warnf("Maximum (%d) number of \"detected_attack\" events exceeded for timeframe: %d / %d ms",
			globals.MaxAttackDetectedEventsPerInterval, len(server.AttackDetectedEventsSentAt), globals.AttackDetectedEventsIntervalInMs)
		return false
	}

	server.AttackDetectedEventsSentAt = append(server.AttackDetectedEventsSentAt, currentTime)
	return true
}

func SendAttackDetectedEvent(server *ServerData, req *protos.AttackDetected) {
	if !ShouldSendAttackDetectedEvent(server) {
		return
	}
	detectedAttackEvent := DetectedAttack{
		Type:    "detected_attack",
		Agent:   GetAgentInfo(server),
		Request: GetRequestInfo(req.Request),
		Attack:  GetAttackDetails(server, req.Attack),
		Time:    utils.GetTime(),
	}

	response, err := SendCloudRequest(server, server.EnvironmentConfig.Endpoint, globals.EventsAPI, globals.EventsAPIMethod, detectedAttackEvent)
	if err != nil {
		LogCloudRequestError(server, "Error in sending detected attack event: ", err)
		return
	}

	StoreCloudConfig(server, response)
}
