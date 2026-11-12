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

func ParseBedrockConverse(body []byte, model string) *protos.AIUsage {
	var response *bedrockConverseResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return nil
	}
	rawUsage := response.Usage
	if !validTokenCounts(rawUsage.Input, rawUsage.Output, rawUsage.CacheRead, rawUsage.CacheWrite) {
		return nil
	}
	// Bedrock's inputTokens excludes cache reads and writes.
	return &protos.AIUsage{
		Model:      bedrockModel(model, response.Model),
		Input:      rawUsage.Input + rawUsage.CacheRead + rawUsage.CacheWrite,
		Output:     rawUsage.Output,
		CacheRead:  rawUsage.CacheRead,
		CacheWrite: rawUsage.CacheWrite,
	}
}

func ParseBedrockInvoke(body []byte, model string) *protos.AIUsage {
	response := decodeAnthropic(body)
	if response == nil || (response.Usage.Input == nil && response.Usage.Output == nil) {
		return nil
	}
	rawUsage := response.Usage
	input, output := response.inputOutputTokens()
	if !validTokenCounts(input, output, rawUsage.CacheRead, rawUsage.CacheWrite) {
		return nil
	}
	return &protos.AIUsage{
		Model:      bedrockModel(model, response.Model),
		Input:      input + rawUsage.CacheRead + rawUsage.CacheWrite,
		Output:     output,
		CacheRead:  rawUsage.CacheRead,
		CacheWrite: rawUsage.CacheWrite,
	}
}

func bedrockModel(model, responseModel string) string {
	model, _ = url.PathUnescape(model)
	if strings.HasPrefix(model, "arn:") {
		return responseModel
	}
	return model
}
