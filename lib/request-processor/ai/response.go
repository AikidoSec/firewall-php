package ai

import (
	"main/ai/providers"
	"main/ipc/protos"
)

const maxBodyBytes = 2 * 1024 * 1024

func ParseResponse(kind ResponseKind, usage *protos.AIUsage, body string) *protos.AIUsage {
	if usage == nil || len(body) > maxBodyBytes {
		return nil
	}
	var parsed bool
	switch kind {
	case openAIChat:
		parsed = providers.ParseOpenAIChat([]byte(body), usage)
	case openAIResponses:
		parsed = providers.ParseOpenAIResponses([]byte(body), usage)
	case anthropic:
		parsed = providers.ParseAnthropic([]byte(body), usage)
	case gemini:
		parsed = providers.ParseGemini([]byte(body), usage)
	case bedrockConverse:
		parsed = providers.ParseBedrockConverse([]byte(body), usage)
	case bedrockInvoke:
		parsed = providers.ParseBedrockInvoke([]byte(body), usage)
	}
	if !parsed || usage.Model == "" || len(usage.Model) > 256 {
		return nil
	}
	if usage.Input < 0 || usage.Output < 0 || usage.CacheRead < 0 || usage.CacheWrite < 0 {
		return nil
	}
	return usage
}
