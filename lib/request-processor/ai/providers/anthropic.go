package providers

import (
	"encoding/json"
	"main/ipc/protos"
)

type anthropicResponse struct {
	Model string `json:"model"`
	Usage *struct {
		Input      *int64 `json:"input_tokens"`
		Output     *int64 `json:"output_tokens"`
		CacheRead  int64  `json:"cache_read_input_tokens"`
		CacheWrite int64  `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

func ParseAnthropic(body []byte) *protos.AIUsage {
	response := decodeAnthropic(body)
	if response == nil {
		return nil
	}
	rawUsage := response.Usage
	input, output := response.inputOutputTokens()
	// Anthropic's input_tokens excludes cache reads and writes.
	return &protos.AIUsage{
		Model:      response.Model,
		Input:      input + rawUsage.CacheRead + rawUsage.CacheWrite,
		Output:     output,
		CacheRead:  rawUsage.CacheRead,
		CacheWrite: rawUsage.CacheWrite,
	}
}

func decodeAnthropic(body []byte) *anthropicResponse {
	var response *anthropicResponse
	if json.Unmarshal(body, &response) != nil || response == nil || response.Usage == nil {
		return nil
	}
	return response
}

func (response *anthropicResponse) inputOutputTokens() (input, output int64) {
	rawUsage := response.Usage
	if rawUsage.Input != nil {
		input = *rawUsage.Input
	}
	if rawUsage.Output != nil {
		output = *rawUsage.Output
	}
	return input, output
}
