package attack

import (
	"encoding/json"
	"fmt"
	"html"
	. "main/aikido_types"
	"main/context"
	"main/globals"
	"main/grpc"
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
func GetHeadersProto() []*protos.Header {
	var headersProto []*protos.Header
	for key, value := range context.GetHeadersParsed() {
		valueStr, ok := value.(string)
		if ok {
			headersProto = append(headersProto, &protos.Header{Key: key, Value: valueStr})
		}
	}
	return headersProto
}

/* Construct the AttackDetected protobuf structure to be sent via gRPC to the Agent */
func GetAttackDetectedProto(server *ServerData, res utils.InterceptorResult) *protos.AttackDetected {
	return &protos.AttackDetected{
		Token: server.AikidoConfig.Token,
		Request: &protos.Request{
			Method:    context.GetMethod(),
			IpAddress: context.GetIp(),
			UserAgent: context.GetUserAgent(),
			Url:       context.GetUrl(),
			Headers:   GetHeadersProto(),
			Body:      context.GetBodyRaw(),
			Source:    "php",
			Route:     context.GetRoute(),
		},
		Attack: &protos.Attack{
			Kind:      string(res.Kind),
			Operation: res.Operation,
			Module:    context.GetModule(),
			Blocked:   utils.IsBlockingEnabled(server),
			Source:    res.Source,
			Path:      res.PathToPayload,
			Stack:     context.GetStackTrace(),
			Payload:   res.Payload,
			Metadata:  GetMetadataProto(res.Metadata),
			UserId:    context.GetUserId(),
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

func ReportAttackDetected(res *utils.InterceptorResult) string {
	if res == nil {
		return ""
	}

	attackDetectedProto := GetAttackDetectedProto(globals.GetCurrentServer(), *res)
	go grpc.OnAttackDetected(attackDetectedProto)
	return GetAttackDetectedAction(*res)
}
