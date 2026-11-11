package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/ipc/protos"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIUsageHeartbeat(t *testing.T) {
	server := &ServerData{}
	server.StoreAIUsage([]*protos.AIUsage{
		{Provider: "https://gateway.test", Model: "gpt-4o", Input: 10, Output: 5, CacheRead: 8},
		{Provider: "https://gateway.test", Model: "gpt-4o", Input: 20, Output: 7, CacheWrite: 4},
		{Provider: "https://other.test", Model: "gpt-4o", Input: 3, Output: 1},
	})
	ai := GetAIUsageAndClear(server)
	require.ElementsMatch(t, []AIStats{
		{Provider: "https://gateway.test", Model: "gpt-4o", Calls: 2, Tokens: AITokens{Input: 30, Output: 12, Total: 42, CacheRead: 8, CacheWrite: 4}},
		{Provider: "https://other.test", Model: "gpt-4o", Calls: 1, Tokens: AITokens{Input: 3, Output: 1, Total: 4}},
	}, ai)
	encoded, err := json.Marshal(Heartbeat{AI: ai})
	require.NoError(t, err)
	var heartbeat map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &heartbeat))
	require.Len(t, heartbeat["ai"], 2)
	require.NotContains(t, heartbeat, "aiTools")
	require.Empty(t, GetAIUsageAndClear(server))
	server.StoreAIUsage([]*protos.AIUsage{{Provider: "https://gateway.test", Model: "gpt-4o", Input: 1}})
	require.Equal(t, int64(1), GetAIUsageAndClear(server)[0].Calls)
}
