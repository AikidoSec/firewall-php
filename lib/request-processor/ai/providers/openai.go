package providers

import (
	"encoding/json"
	"main/ipc/protos"
)

type openAICacheDetails struct {
	CacheRead  int64 `json:"cached_tokens"`
	CacheWrite int64 `json:"cache_write_tokens"`
}

type openAIChatResponse struct {
	Model string `json:"model"`
	Usage *struct {
		Input   int64              `json:"prompt_tokens"`
		Output  int64              `json:"completion_tokens"`
		Details openAICacheDetails `json:"prompt_tokens_details"`
	} `json:"usage"`
}

type openAIResponse struct {
	Model string `json:"model"`
	Usage *struct {
		Input   int64              `json:"input_tokens"`
		Output  int64              `json:"output_tokens"`
		Details openAICacheDetails `json:"input_tokens_details"`
	} `json:"usage"`
}

func ParseOpenAIChat(body []byte, usage *protos.AIUsage) bool {
	var response *openAIChatResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return false
	}
	rawUsage := response.Usage
	usage.Model = response.Model
	usage.Input = rawUsage.Input
	usage.Output = rawUsage.Output
	usage.CacheRead = rawUsage.Details.CacheRead
	usage.CacheWrite = rawUsage.Details.CacheWrite
	return true
}

func ParseOpenAIResponses(body []byte, usage *protos.AIUsage) bool {
	var response *openAIResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return false
	}
	rawUsage := response.Usage
	usage.Model = response.Model
	usage.Input = rawUsage.Input
	usage.Output = rawUsage.Output
	usage.CacheRead = rawUsage.Details.CacheRead
	usage.CacheWrite = rawUsage.Details.CacheWrite
	return true
}
