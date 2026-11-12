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

func ParseGemini(body []byte, model string) *protos.AIUsage {
	var response *geminiResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return nil
	}
	rawUsage := response.Usage
	if !validTokenCounts(rawUsage.Input, rawUsage.ToolInput, rawUsage.Output, rawUsage.Thoughts, rawUsage.CacheRead) {
		return nil
	}
	finalModel := model
	if response.Model != "" {
		finalModel = response.Model
	}
	return &protos.AIUsage{
		Model:     finalModel,
		Input:     rawUsage.Input + rawUsage.ToolInput,
		Output:    rawUsage.Output + rawUsage.Thoughts,
		CacheRead: rawUsage.CacheRead,
	}
}
