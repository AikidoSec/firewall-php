package grpc

import (
	"context"
	"main/ipc/protos"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	transportgrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type aiUsageTestServer struct {
	protos.UnimplementedAikidoServer
	received chan *protos.AIUsage
}

func (s *aiUsageTestServer) OnAIUsage(ctx context.Context, req *protos.AIUsage) (*protos.Empty, error) {
	select {
	case s.received <- req:
		return &protos.Empty{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestAIUsageReportingOverRPC(t *testing.T) {
	oldClient := client
	t.Cleanup(func() { client = oldClient })
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	server := &aiUsageTestServer{received: make(chan *protos.AIUsage, 3)}
	transport := transportgrpc.NewServer()
	protos.RegisterAikidoServer(transport, server)
	go transport.Serve(listener)
	t.Cleanup(transport.Stop)
	connection, err := dialAgent(listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })
	client = protos.NewAikidoClient(connection)
	usages := []*protos.AIUsage{
		{Token: "first", ServerPid: 77, Provider: "https://gateway.test", Model: "gpt", Input: 10, Output: 5},
		{Token: "second", ServerPid: 77, Provider: "https://gateway.test", Model: "claude", Input: 20},
		{Token: "first", ServerPid: 78, Provider: "https://other.test", Model: "gpt", Input: 7},
	}
	for _, usage := range usages {
		OnAIUsage(usage)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, expected := range usages {
		select {
		case actual := <-server.received:
			require.True(t, proto.Equal(expected, actual))
		case <-ctx.Done():
			t.Fatal("timed out waiting for AI usage")
		}
	}
}
