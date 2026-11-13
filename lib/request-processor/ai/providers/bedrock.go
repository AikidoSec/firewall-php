package providers

import (
	"encoding/json"
	"main/ipc/protos"
	"net/url"
	"strings"
)

type bedrockConverseResponse struct {
	Model string `json:"model"`
	Usage *struct {
		Input      int64 `json:"inputTokens"`
		Output     int64 `json:"outputTokens"`
		CacheRead  int64 `json:"cacheReadInputTokens"`
		CacheWrite int64 `json:"cacheWriteInputTokens"`
	} `json:"usage"`
}

func ParseBedrockConverse(body []byte, usage *protos.AIUsage) bool {
	var response *bedrockConverseResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return false
	}
	rawUsage := response.Usage
	// Bedrock's inputTokens excludes cache reads and writes.
	usage.Model = bedrockModel(usage.Model, response.Model)
	usage.Input = rawUsage.Input + rawUsage.CacheRead + rawUsage.CacheWrite
	usage.Output = rawUsage.Output
	usage.CacheRead = rawUsage.CacheRead
	usage.CacheWrite = rawUsage.CacheWrite
	return true
}

func ParseBedrockInvoke(body []byte, usage *protos.AIUsage) bool {
	response := decodeAnthropic(body)
	if response == nil || (response.Usage.Input == nil && response.Usage.Output == nil) {
		return false
	}
	rawUsage := response.Usage
	input, output := response.inputOutputTokens()
	usage.Model = bedrockModel(usage.Model, response.Model)
	usage.Input = input + rawUsage.CacheRead + rawUsage.CacheWrite
	usage.Output = output
	usage.CacheRead = rawUsage.CacheRead
	usage.CacheWrite = rawUsage.CacheWrite
	return true
}

func bedrockModel(model, responseModel string) string {
	model, _ = url.PathUnescape(model)
	if strings.HasPrefix(model, "arn:") && responseModel != "" {
		return responseModel
	}
	return model
}
