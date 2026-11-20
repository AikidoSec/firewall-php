package providers

import (
	"encoding/json"
	"main/ipc/protos"
)

type geminiResponse struct {
	Model string `json:"modelVersion"`
	Usage *struct {
		Input     int64 `json:"promptTokenCount"`
		ToolInput int64 `json:"toolUsePromptTokenCount"`
		Output    int64 `json:"candidatesTokenCount"`
		Thoughts  int64 `json:"thoughtsTokenCount"`
		CacheRead int64 `json:"cachedContentTokenCount"`
	} `json:"usageMetadata"`
}

type geminiInteractionResponse struct {
	Model string `json:"model"`
	Agent string `json:"agent"`
	Usage *struct {
		Input     int64 `json:"total_input_tokens"`
		ToolInput int64 `json:"total_tool_use_tokens"`
		Output    int64 `json:"total_output_tokens"`
		Thoughts  int64 `json:"total_thought_tokens"`
		CacheRead int64 `json:"total_cached_tokens"`
	} `json:"usage"`
}

func ParseGemini(body []byte, usage *protos.AIUsage) bool {
	var response *geminiResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return false
	}
	rawUsage := response.Usage
	if response.Model != "" {
		usage.Model = response.Model
	}
	usage.Input = rawUsage.Input + rawUsage.ToolInput
	usage.Output = rawUsage.Output + rawUsage.Thoughts
	usage.CacheRead = rawUsage.CacheRead
	return true
}

func ParseGeminiInteraction(body []byte, usage *protos.AIUsage) bool {
	var response *geminiInteractionResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return false
	}
	rawUsage := response.Usage
	usage.Model = response.Model
	if usage.Model == "" {
		usage.Model = response.Agent
	}
	usage.Input = rawUsage.Input + rawUsage.ToolInput
	usage.Output = rawUsage.Output + rawUsage.Thoughts
	usage.CacheRead = rawUsage.CacheRead
	return true
}
