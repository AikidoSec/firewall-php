package grpc

import (
	"context"
	"main/globals"
	"main/ipc/protos"
	"main/utils"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type largeCloudConfigServer struct {
	protos.UnimplementedAikidoServer
	config *protos.CloudConfig
}

func (s *largeCloudConfigServer) GetCloudConfig(context.Context, *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	return s.config, nil
}

func TestClientReceivesAndAppliesLargeCloudConfig(t *testing.T) {
	ips := make([]string, 300000)
	for i := range ips {
		ips[i] = "2001:0db8:0000:0000:0000:0000:0000:0001/128"
	}
	config := &protos.CloudConfig{
		ConfigUpdatedAt: 1,
		Block:           true,
		BlockedIps: map[string]*protos.IpList{
			"geoip/test": {Description: "Geo restrictions", Ips: []string{"198.51.100.0/24"}},
		},
		MonitoredIps: map[string]*protos.IpList{"large-list": {Ips: ips}},
	}
	require.Greater(t, proto.Size(config), 10*1024*1024)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	transport := grpc.NewServer(grpc.MaxSendMsgSize(maxMessageSize))
	protos.RegisterAikidoServer(transport, &largeCloudConfigServer{config: config})
	go transport.Serve(listener)
	t.Cleanup(transport.Stop)
	connection, err := dialAgent(listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })
	previousClient := client
	client = protos.NewAikidoClient(connection)
	t.Cleanup(func() { client = previousClient })

	worker := globals.NewServerData()
	worker.AikidoConfig.Token = "AIK_LARGE_CONFIG_TEST"
	GetCloudConfig(worker, 10*time.Second)
	require.Equal(t, int64(1), worker.CloudConfig.ConfigUpdatedAt)
	require.Equal(t, 1, worker.CloudConfig.Block)
	blocked, _ := utils.IsIpBlocked(nil, worker, "198.51.100.42")
	require.True(t, blocked, "the geo rule must be enforced after receiving the large config")
	monitored, _ := utils.IsIpMonitored(nil, worker, "2001:db8::1")
	require.True(t, monitored, "the large list must also be applied")
}
