package aikido_types

import (
	"fmt"
	"main/ipc/protos"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreAIUsageIgnoresIncompleteCalls(t *testing.T) {
	server := &ServerData{}
	for _, call := range []*protos.AIUsage{nil, {}, {Provider: "https://gateway.test"}, {Model: "gpt"}} {
		server.StoreAIUsage(call)
	}
	require.Empty(t, server.AI)
}

func TestStoreAIUsageBoundsDistinctModels(t *testing.T) {
	server := &ServerData{}
	for i := 0; i < MaxNumberOfStoredAIStats+10; i++ {
		server.StoreAIUsage(&protos.AIUsage{Provider: "https://gateway.test", Model: fmt.Sprint(i), Input: 1})
	}
	require.Len(t, server.AI, MaxNumberOfStoredAIStats)
	server.StoreAIUsage(&protos.AIUsage{Provider: "https://gateway.test", Model: "0", Input: 7})
	stats := server.AI[AIKey{Provider: "https://gateway.test", Model: "0"}]
	require.Equal(t, int64(2), stats.Calls)
	require.Equal(t, int64(8), stats.Tokens.Input)
	require.NotContains(t, server.AI, AIKey{Provider: "https://gateway.test", Model: fmt.Sprint(MaxNumberOfStoredAIStats)})
}
