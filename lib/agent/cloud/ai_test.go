package cloud

import (
	"encoding/json"
	. "main/aikido_types"
	"main/ipc/protos"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIUsageHeartbeat(t *testing.T) {
	server := &ServerData{}
	for _, call := range []*protos.AIUsage{
		{Provider: "https://gateway.test", Model: "gpt-4o", Input: 10, Output: 5},
		{Provider: "https://gateway.test", Model: "gpt-4o", Input: 20, Output: 7},
		{Provider: "https://other.test", Model: "gpt-4o", Input: 3, Output: 1},
		{Provider: "https://gateway.test", Model: "claude", Input: 4, Output: 2},
	} {
		server.StoreAIUsage(call)
	}
	ai := GetAIUsageAndClear(server)
	require.ElementsMatch(t, []AIStats{
		{Provider: "https://gateway.test", Model: "gpt-4o", Calls: 2, Tokens: AITokens{Input: 30, Output: 12, Total: 42}},
		{Provider: "https://other.test", Model: "gpt-4o", Calls: 1, Tokens: AITokens{Input: 3, Output: 1, Total: 4}},
		{Provider: "https://gateway.test", Model: "claude", Calls: 1, Tokens: AITokens{Input: 4, Output: 2, Total: 6}},
	}, ai)
	encoded, err := json.Marshal(Heartbeat{AI: ai})
	require.NoError(t, err)
	var heartbeat map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &heartbeat))
	require.Len(t, heartbeat["ai"], 3)
	require.Empty(t, GetAIUsageAndClear(server))
	server.StoreAIUsage(&protos.AIUsage{Provider: "https://gateway.test", Model: "gpt-4o", Input: 1})
	require.Equal(t, int64(1), GetAIUsageAndClear(server)[0].Calls)
}

func TestAIUsageConcurrentCollection(t *testing.T) {
	server := &ServerData{}
	const workers, callsPerWorker = 4, 200
	var writers sync.WaitGroup
	for i := 0; i < workers; i++ {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for j := 0; j < callsPerWorker; j++ {
				server.StoreAIUsage(&protos.AIUsage{Provider: "https://gateway.test", Model: "gpt", Input: 10, Output: 3})
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		writers.Wait()
		close(done)
	}()
	var calls, input, output int64
	collect := func() {
		for _, stats := range GetAIUsageAndClear(server) {
			calls += stats.Calls
			input += stats.Tokens.Input
			output += stats.Tokens.Output
		}
	}
	for {
		collect()
		select {
		case <-done:
			collect()
			require.Equal(t, int64(workers*callsPerWorker), calls)
			require.Equal(t, calls*10, input)
			require.Equal(t, calls*3, output)
			return
		default:
			runtime.Gosched()
		}
	}
}
