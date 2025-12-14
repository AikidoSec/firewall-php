package attack

import (
	"encoding/json"
	"fmt"
	"html"
	"main/context"
	"main/grpc"
	"main/instance"
	"main/ipc/protos"
	"main/utils"
)

/* Convert metadata map to protobuf structure to be sent via gRPC to the Agent */
func GetMetadataProto(metadata map[string]string) []*protos.Metadata {
	var metadataProto []*protos.Metadata
	for key, value := range metadata {
		metadataProto = append(metadataProto, &protos.Metadata{Key: key, Value: value})
	}
	return metadataProto
}

/* Convert headers map to protobuf structure to be sent via gRPC to the Agent */
func GetHeadersProto(inst *instance.RequestProcessorInstance) []*protos.Header {
	var headersProto []*protos.Header
	for key, value := range context.GetHeadersParsed(inst) {
		valueStr, ok := value.(string)
		if ok {
			headersProto = append(headersProto, &protos.Header{Key: key, Value: valueStr})
		}
	}
	return headersProto
}

/* Construct the AttackDetected protobuf structure to be sent via gRPC to the Agent */
func GetAttackDetectedProto(res utils.InterceptorResult, inst *instance.RequestProcessorInstance) *protos.AttackDetected {
	token := inst.GetCurrentToken()
	server := inst.GetCurrentServer()

	serverPID := context.GetServerPID()
	return &protos.AttackDetected{
		Token:     token,
		ServerPid: serverPID,
		Request: &protos.Request{
			Method:    context.GetMethod(inst),
			IpAddress: context.GetIp(inst),
			UserAgent: context.GetUserAgent(inst),
			Url:       context.GetUrl(inst),
			Headers:   GetHeadersProto(inst),
			Body:      context.GetBodyRaw(inst),
			Source:    "php",
			Route:     context.GetRoute(inst),
		},
		Attack: &protos.Attack{
			Kind:      string(res.Kind),
			Operation: res.Operation,
			Module:    context.GetModule(inst),
			Blocked:   utils.IsBlockingEnabled(server),
			Source:    res.Source,
			Path:      res.PathToPayload,
			Stack:     context.GetStackTrace(inst),
			Payload:   res.Payload,
			Metadata:  GetMetadataProto(res.Metadata),
			UserId:    context.GetUserId(inst),
		},
	}
}

func BuildAttackDetectedMessage(result utils.InterceptorResult) string {
	return fmt.Sprintf("Aikido firewall has blocked %s: %s(...) originating from %s%s",
		utils.GetDisplayNameForAttackKind(result.Kind),
		result.Operation,
		result.Source,
		html.EscapeString(result.PathToPayload))
}

func GetThrowAction(message string, code int) string {
	actionMap := map[string]interface{}{
		"action":  "throw",
		"message": message,
		"code":    code,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}

func GetAttackDetectedAction(result utils.InterceptorResult) string {
	return GetThrowAction(BuildAttackDetectedMessage(result), 500)
}

func ReportAttackDetected(res *utils.InterceptorResult, inst *instance.RequestProcessorInstance) string {
	if res == nil {
		return ""
	}

	grpc.OnAttackDetected(inst, GetAttackDetectedProto(*res, inst))
	return GetAttackDetectedAction(*res)
}
