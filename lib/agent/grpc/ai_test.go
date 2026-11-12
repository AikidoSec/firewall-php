package grpc

import (
	"context"
	. "main/aikido_types"
	"main/cloud"
	"main/globals"
	"main/ipc/protos"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAIUsageRPCIsolatesTokensAndServerProcesses(t *testing.T) {
	keys := []ServerKey{
		{Token: "AIK_AI_USAGE_FIRST", ServerPID: 1},
		{Token: "AIK_AI_USAGE_FIRST", ServerPID: 2},
		{Token: "AIK_AI_USAGE_SECOND", ServerPID: 1},
	}
	for _, key := range keys {
		globals.ServersMutex.Lock()
		globals.Servers[key] = &ServerData{}
		globals.ServersMutex.Unlock()
	}
	t.Cleanup(func() {
		globals.ServersMutex.Lock()
		defer globals.ServersMutex.Unlock()
		for _, key := range keys {
			delete(globals.Servers, key)
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	transport := newServer()
	go transport.Serve(listener)
	t.Cleanup(transport.Stop)
	connection, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })
	client := protos.NewAikidoClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i, key := range keys {
		_, err := client.OnAIUsage(ctx, &protos.AIUsage{
			Token: key.Token, ServerPid: key.ServerPID,
			Provider: "https://gateway.test", Model: "gpt", Input: int64(i + 1),
		})
		require.NoError(t, err)
	}
	unknown := ServerKey{Token: keys[0].Token, ServerPID: 99}
	_, err = client.OnAIUsage(ctx, &protos.AIUsage{
		Token: unknown.Token, ServerPid: unknown.ServerPID,
		Provider: "https://gateway.test", Model: "gpt", Input: 100,
	})
	require.NoError(t, err)
	require.Nil(t, globals.GetServer(unknown))
	for i, key := range keys {
		stats := cloud.GetAIUsageAndClear(globals.GetServer(key))
		require.Len(t, stats, 1)
		require.Equal(t, int64(1), stats[0].Calls)
		require.Equal(t, int64(i+1), stats[0].Tokens.Input)
	}
}
