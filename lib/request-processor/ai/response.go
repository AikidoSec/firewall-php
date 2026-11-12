package ai

import (
	"main/ai/providers"
	"main/ipc/protos"
	"strings"
)

const maxBodyBytes = 2 * 1024 * 1024

func SupportsURL(rawURL string) bool {
	return matchEndpoint(rawURL).kind != unknownEndpoint
}

func ParseResponse(rawURL, body string) *protos.AIUsage {
	if len(body) > maxBodyBytes {
		return nil
	}
	endpoint := matchEndpoint(rawURL)
	var usage *protos.AIUsage
	switch endpoint.kind {
	case openAIChat:
		usage = providers.ParseOpenAIChat([]byte(body))
	case openAIResponses:
		usage = providers.ParseOpenAIResponses([]byte(body))
	case anthropic:
		usage = providers.ParseAnthropic([]byte(body))
	case gemini:
		usage = providers.ParseGemini([]byte(body), endpoint.model)
	case bedrockConverse:
		usage = providers.ParseBedrockConverse([]byte(body), endpoint.model)
	case bedrockInvoke:
		usage = providers.ParseBedrockInvoke([]byte(body), endpoint.model)
	}
	if usage == nil || usage.Model == "" || len(usage.Model) > 256 {
		return nil
	}
	// URL-derived models must not retain the caller's URL buffer.
	usage.Model = strings.Clone(usage.Model)
	usage.Provider = endpoint.origin
	return usage
}
