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

func GetAttackDetails(protoAttack *protos.Attack) AttackDetails {
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
		User:      utils.GetUserById(protoAttack.UserId),
	}
}

func ShouldSendAttackDetectedEvent() bool {
	globals.AttackDetectedEventsSentAtMutex.Lock()
	defer globals.AttackDetectedEventsSentAtMutex.Unlock()

	currentTime := utils.GetTime()

	// Filter out events that are outside the current interval
	var filteredEvents []int64
	for _, eventTime := range globals.AttackDetectedEventsSentAt {
		if eventTime > currentTime-globals.AttackDetectedEventsIntervalInMs {
			filteredEvents = append(filteredEvents, eventTime)
		}
	}
	globals.AttackDetectedEventsSentAt = filteredEvents

	if len(globals.AttackDetectedEventsSentAt) >= globals.MaxAttackDetectedEventsPerInterval {
		log.Warnf("Maximum (%d) number of \"detected_attack\" events exceeded for timeframe: %d / %d ms",
			globals.MaxAttackDetectedEventsPerInterval, len(globals.AttackDetectedEventsSentAt), globals.AttackDetectedEventsIntervalInMs)
		return false
	}

	globals.AttackDetectedEventsSentAt = append(globals.AttackDetectedEventsSentAt, currentTime)
	return true
}

func SendAttackDetectedEvent(req *protos.AttackDetected) {
	if !ShouldSendAttackDetectedEvent() {
		return
	}
	detectedAttackEvent := DetectedAttack{
		Type:    "detected_attack",
		Agent:   GetAgentInfo(),
		Request: GetRequestInfo(req.Request),
		Attack:  GetAttackDetails(req.Attack),
		Time:    utils.GetTime(),
	}

	response, err := SendCloudRequest(globals.EnvironmentConfig.Endpoint, globals.EventsAPI, globals.EventsAPIMethod, detectedAttackEvent)
	if err != nil {
		LogCloudRequestError("Error in sending detected attack event: ", err)
		return
	}

	StoreCloudConfig(response)
}
