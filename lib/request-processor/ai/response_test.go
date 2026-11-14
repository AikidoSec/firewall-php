package ai

import (
	"fmt"
	"main/ipc/protos"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func parseResponseForURL(rawURL, body string) *protos.AIUsage {
	kind, usage := ClassifyURL(rawURL)
	return ParseResponse(kind, usage, body)
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name, path, body string
		kind             ResponseKind
		want             *protos.AIUsage
	}{
		{"OpenAI chat", "/custom/chat/completions?key=secret", `{"model":"gpt","usage":{"prompt_tokens":10,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":3,"cache_write_tokens":2}}}`, openAIChat, &protos.AIUsage{Model: "gpt", Input: 10, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"OpenAI responses", "/v1/responses", `{"model":"gpt","usage":{"input_tokens":10,"output_tokens":5,"input_tokens_details":{"cached_tokens":3,"cache_write_tokens":2}}}`, openAIResponses, &protos.AIUsage{Model: "gpt", Input: 10, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"Anthropic", "/v1/messages", `{"model":"claude","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":3,"cache_creation_input_tokens":2}}`, anthropic, &protos.AIUsage{Model: "claude", Input: 15, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"Anthropic cached input", "/v1/messages", `{"model":"claude","usage":{"input_tokens":0,"output_tokens":5,"cache_read_input_tokens":100}}`, anthropic, &protos.AIUsage{Model: "claude", Input: 100, Output: 5, CacheRead: 100}},
		{"Vertex Anthropic", "/publishers/anthropic/models/claude:rawPredict", `{"model":"claude","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":3,"cache_creation_input_tokens":2}}`, anthropic, &protos.AIUsage{Model: "claude", Input: 15, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"Gemini", "/v1beta/models/gemini:generateContent", `{"modelVersion":"gemini-2","usageMetadata":{"promptTokenCount":8,"toolUsePromptTokenCount":2,"candidatesTokenCount":3,"thoughtsTokenCount":2,"cachedContentTokenCount":4}}`, gemini, &protos.AIUsage{Model: "gemini-2", Input: 10, Output: 5, CacheRead: 4}},
		{"Gemini model fallback", "/models/gemini:generateContent", `{"usageMetadata":{"promptTokenCount":10}}`, gemini, &protos.AIUsage{Model: "gemini", Input: 10}},
		{"Bedrock Converse", "/model/anthropic.claude/converse", `{"usage":{"inputTokens":10,"outputTokens":5,"cacheReadInputTokens":3,"cacheWriteInputTokens":2}}`, bedrockConverse, &protos.AIUsage{Model: "anthropic.claude", Input: 15, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"Bedrock Invoke", "/model/anthropic.claude/invoke", `{"usage":{"input_tokens":10,"output_tokens":5}}`, bedrockInvoke, &protos.AIUsage{Model: "anthropic.claude", Input: 10, Output: 5}},
		{"Bedrock Invoke cached input", "/model/anthropic.claude/invoke", `{"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":3,"cache_creation_input_tokens":2}}`, bedrockInvoke, &protos.AIUsage{Model: "anthropic.claude", Input: 15, Output: 5, CacheRead: 3, CacheWrite: 2}},
		{"Bedrock encoded model", "/model/anthropic%2Eclaude/invoke", `{"usage":{"input_tokens":10}}`, bedrockInvoke, &protos.AIUsage{Model: "anthropic.claude", Input: 10}},
		{"Bedrock ARN model", "/model/arn%3Aaws%3Abedrock%3Aregion%3Aaccount%3Amodel%2Fname/invoke", `{"model":"claude","usage":{"input_tokens":10}}`, bedrockInvoke, &protos.AIUsage{Model: "claude", Input: 10}},
		{"Bedrock Converse ARN without response model", "/model/arn%3Aaws%3Abedrock%3Aus-east-1%3A123456789012%3Ainference-profile%2Fmy-profile/converse", `{"usage":{"inputTokens":10,"outputTokens":5}}`, bedrockConverse, &protos.AIUsage{Model: "arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile", Input: 10, Output: 5}},
		{"Bedrock Invoke ARN without response model", "/model/arn%3Aaws%3Abedrock%3Aus-east-1%3A123456789012%3Ainference-profile%2Fmy-profile/invoke", `{"usage":{"input_tokens":10,"output_tokens":5}}`, bedrockInvoke, &protos.AIUsage{Model: "arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile", Input: 10, Output: 5}},
		{"nesting characters in text", "/v1/responses", `{"model":"gpt","usage":{},"text":"[[[\"\\}}"}`, openAIResponses, &protos.AIUsage{Model: "gpt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Provider = "gateway.test"
			kind, usage := ClassifyURL("https://user:password@GATEWAY.test:443" + tt.path)
			require.NotNil(t, usage)
			require.Equal(t, tt.kind, kind)
			parsed := ParseResponse(kind, usage, tt.body)
			require.Same(t, usage, parsed)
			require.Equal(t, tt.want, parsed)
		})
	}
}

func TestParseResponseRejectsInvalidResponses(t *testing.T) {
	for name, body := range map[string]string{
		"invalid JSON": "{invalid", "trailing JSON": `{"model":"gpt"}{}`, "array": `[]`, "null": `null`,
		"streaming response": "data: {\"model\":\"gpt\"}\n\ndata: [DONE]\n\n",
		"missing model":      `{"usage":{}}`, "non-string model": `{"model":123,"usage":{}}`,
		"missing usage": `{"model":"gpt"}`, "null usage": `{"model":"gpt","usage":null}`,
		"long model":    `{"model":"` + strings.Repeat("x", 257) + `","usage":{}}`,
		"oversized":     `{"model":"gpt","usage":{},"text":"` + strings.Repeat("x", maxBodyBytes) + `"}`,
		"deeply nested": `{"model":"gpt","usage":{},"data":` + strings.Repeat("[", 10001) + "0" + strings.Repeat("]", 10001) + "}",
	} {
		t.Run(name, func(t *testing.T) {
			require.Nil(t, parseResponseForURL("https://gateway.test/v1/responses", body))
		})
	}
	require.Nil(t, parseResponseForURL("https://gateway.test/model/claude/invoke", `{"model":"claude"}`))
	require.Nil(t, parseResponseForURL("https://gateway.test/model/claude/invoke", `{"usage":{"prompt_tokens":10}}`))
}

func TestParseResponseRejectsMalformedUsage(t *testing.T) {
	for _, tt := range []struct {
		name, path, body string
	}{
		{"OpenAI chat details", "/v1/chat/completions", `{"model":"gpt","usage":{"prompt_tokens":10,"prompt_tokens_details":[]}}`},
		{"OpenAI responses input", "/v1/responses", `{"model":"gpt","usage":{"input_tokens":"10"}}`},
		{"Anthropic cache", "/v1/messages", `{"model":"claude","usage":{"input_tokens":10,"cache_read_input_tokens":"3"}}`},
		{"Bedrock Converse output", "/model/claude/converse", `{"usage":{"inputTokens":10,"outputTokens":false}}`},
		{"Bedrock Invoke usage", "/model/claude/invoke", `{"usage":[]}`},
		{"Gemini thoughts", "/models/gemini:generateContent", `{"usageMetadata":{"promptTokenCount":10,"thoughtsTokenCount":"5"}}`},
		{"Bedrock null", "/model/claude/converse", `null`},
		{"Gemini null", "/models/gemini:generateContent", `null`},
		{"Gemini missing usage", "/models/gemini:generateContent", `{}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Nil(t, parseResponseForURL("https://gateway.test"+tt.path, tt.body))
		})
	}
}

func TestUnsupportedURLs(t *testing.T) {
	for _, rawURL := range []string{
		"https://gateway.test/other", "ftp://gateway.test/v1/responses", "/v1/responses", "https:///v1/responses",
		"https://gateway.test/%zz/responses", "https://gateway.test/publishers/other/models/claude:rawPredict",
		"https://gateway.test/models/foo/bar:generateContent", "https://gateway.test/model/foo/bar/converse",
		"https://gateway.test/v1/responses/", "https://gateway.test/other?next=/v1/responses",
	} {
		t.Run(rawURL, func(t *testing.T) {
			kind, usage := ClassifyURL(rawURL)
			require.Nil(t, usage)
			require.Nil(t, ParseResponse(kind, usage, `{"model":"gpt","usage":{}}`))
		})
	}
}

func TestProviderHostname(t *testing.T) {
	for rawURL, hostname := range map[string]string{
		"HTTP://Example.test:80": "example.test",
		"http://[::1]:8080":      "::1",
	} {
		t.Run(rawURL, func(t *testing.T) {
			call := parseResponseForURL(rawURL+"/v1/responses?key=secret#fragment", `{"model":"gpt","usage":{}}`)
			require.NotNil(t, call)
			require.Equal(t, hostname, call.Provider)
		})
	}
}

func TestClassifyURLOwnsDerivedStrings(t *testing.T) {
	for _, tt := range []struct{ path, body string }{
		{"/model/claude/converse", `{"usage":{"inputTokens":1}}`},
		{"/model/claude/invoke", `{"usage":{"input_tokens":1}}`},
		{"/models/claude:generateContent", `{"usageMetadata":{"promptTokenCount":1}}`},
	} {
		t.Run(tt.path, func(t *testing.T) {
			buffer := []byte("https://gateway.test" + tt.path)
			// cgo strings borrow the caller's buffer only for the duration of the call.
			borrowedURL := unsafe.String(unsafe.SliceData(buffer), len(buffer))
			kind, usage := ClassifyURL(borrowedURL)
			clear(buffer)
			call := ParseResponse(kind, usage, tt.body)
			require.NotNil(t, call)
			require.Equal(t, "claude", call.Model)
			require.Equal(t, "gateway.test", call.Provider)
		})
	}
}

func TestTokenCounts(t *testing.T) {
	for _, value := range []string{"-1", "1.5", "1e2", `"10"`, "true", "{}", "[]", "9223372036854775808", "18446744073709551615"} {
		t.Run(value, func(t *testing.T) {
			body := fmt.Sprintf(`{"model":"gpt","usage":{"input_tokens":%s,"output_tokens":5}}`, value)
			call := parseResponseForURL("https://gateway.test/v1/responses", body)
			require.Nil(t, call)
		})
	}
	for _, value := range []int64{1 << 53, math.MaxInt64} {
		body := fmt.Sprintf(`{"model":"gpt","usage":{"input_tokens":%d}}`, value)
		call := parseResponseForURL("https://gateway.test/v1/responses", body)
		require.NotNil(t, call)
		require.Equal(t, value, call.Input)
	}
	call := parseResponseForURL("https://gateway.test/v1/responses", `{"model":"gpt","usage":{"input_tokens":null,"output_tokens":5}}`)
	require.NotNil(t, call)
	require.Zero(t, call.Input)
	require.Equal(t, int64(5), call.Output)
}
