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

func ParseOpenAIChat(body []byte) *protos.AIUsage {
	var response *openAIChatResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return nil
	}
	rawUsage := response.Usage
	if !validTokenCounts(rawUsage.Input, rawUsage.Output, rawUsage.Details.CacheRead, rawUsage.Details.CacheWrite) {
		return nil
	}
	return &protos.AIUsage{
		Model:      response.Model,
		Input:      rawUsage.Input,
		Output:     rawUsage.Output,
		CacheRead:  rawUsage.Details.CacheRead,
		CacheWrite: rawUsage.Details.CacheWrite,
	}
}

func ParseOpenAIResponses(body []byte) *protos.AIUsage {
	var response *openAIResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return nil
	}
	rawUsage := response.Usage
	if !validTokenCounts(rawUsage.Input, rawUsage.Output, rawUsage.Details.CacheRead, rawUsage.Details.CacheWrite) {
		return nil
	}
	return &protos.AIUsage{
		Model:      response.Model,
		Input:      rawUsage.Input,
		Output:     rawUsage.Output,
		CacheRead:  rawUsage.Details.CacheRead,
		CacheWrite: rawUsage.Details.CacheWrite,
	}
}
