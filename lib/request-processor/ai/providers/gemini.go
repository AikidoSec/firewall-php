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
