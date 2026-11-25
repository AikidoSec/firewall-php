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
func GetHeadersProto(instance *instance.RequestProcessorInstance) []*protos.Header {
	var headersProto []*protos.Header
	for key, value := range context.GetHeadersParsed(instance) {
		valueStr, ok := value.(string)
		if ok {
			headersProto = append(headersProto, &protos.Header{Key: key, Value: valueStr})
		}
	}
	return headersProto
}

/* Construct the AttackDetected protobuf structure to be sent via gRPC to the Agent */
func GetAttackDetectedProto(res utils.InterceptorResult, instance *instance.RequestProcessorInstance) *protos.AttackDetected {
	token := instance.GetCurrentToken()
	server := instance.GetCurrentServer()

	serverPID := context.GetServerPID()
	request := &protos.Request{Source: "php", Method: context.GetMethod(instance)}
	if request.Method != "" {
		request.IpAddress = context.GetIp(instance)
		request.UserAgent = context.GetUserAgent(instance)
		request.Url = context.GetUrl(instance)
		request.Headers = GetHeadersProto(instance)
		request.Body = context.GetBodyRaw(instance)
		request.Route = context.GetRoute(instance)
	}
	return &protos.AttackDetected{
		Token:     token,
		ServerPid: serverPID,
		Request:   request,
		Attack: &protos.Attack{
			Kind:      string(res.Kind),
			Operation: res.Operation,
			Module:    context.GetModule(instance),
			Blocked:   utils.IsBlockingEnabled(server),
			Source:    res.Source,
			Path:      res.PathToPayload,
			Stack:     context.GetStackTrace(instance),
			Payload:   res.Payload,
			Metadata:  GetMetadataProto(res.Metadata),
			UserId:    context.GetUserId(instance),
		},
	}
}

func BuildAttackDetectedMessage(result utils.InterceptorResult) string {
	message := fmt.Sprintf("Aikido firewall has blocked %s: %s(...)", utils.GetDisplayNameForAttackKind(result.Kind), result.Operation)
	if result.Source != "" {
		message += " originating from " + result.Source + html.EscapeString(result.PathToPayload)
	}
	return message
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

func ReportAttackDetected(res *utils.InterceptorResult, instance *instance.RequestProcessorInstance) string {
	if res == nil {
		return ""
	}

	grpc.OnAttackDetected(instance, GetAttackDetectedProto(*res, instance))
	return GetAttackDetectedAction(*res)
}
