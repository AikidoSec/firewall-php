package grpc

import (
	"context"
	"fmt"
	"main/aikido_types"
	"main/cloud"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/rate_limiting"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcServer struct {
	protos.AikidoServer
}

func createNewServer(token string) *aikido_types.ServerData {
	globals.ServersMutex.Lock()
	defer globals.ServersMutex.Unlock()
	globals.Servers[token] = aikido_types.NewServerData()
	return globals.Servers[token]
}

func (s *GrpcServer) OnConfig(ctx context.Context, req *protos.Config) (*emptypb.Empty, error) {
	log.Infof("OnConfig called with token: %s", req.GetToken())
	token := req.GetToken()
	if token == "" {
		return &emptypb.Empty{}, nil
	}

	server := globals.GetServer(token)
	if server != nil {
		log.Debugf("Server %s already exists, skipping config update...", token)
		return &emptypb.Empty{}, nil
	}
	log.Debugf("Storing new server %s, initializing cloud, rate limiting and storing config...", token)

	server = createNewServer(token)
	storeConfig(server, req)

	cloud.InitServer(server)
	rate_limiting.InitServer(server)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnPackages(ctx context.Context, req *protos.Packages) (*emptypb.Empty, error) {
	log.Infof("OnPackages called with token: %s", req.GetToken())
	storePackages(globals.GetServer(req.GetToken()), req.GetPackages())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnDomain(ctx context.Context, req *protos.Domain) (*emptypb.Empty, error) {
	log.Infof("OnDomain called with token: %s", req.GetToken())
	log.Debugf("Received domain: %s:%d", req.GetDomain(), req.GetPort())
	storeDomain(globals.GetServer(req.GetToken()), req.GetDomain(), req.GetPort())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) GetRateLimitingStatus(ctx context.Context, req *protos.RateLimitingInfo) (*protos.RateLimitingStatus, error) {
	log.Infof("GetRateLimitingStatus called with token: %s", req.GetToken())
	log.Debugf("Received rate limiting info: %s %s %s %s %s %s", req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())

	return getRateLimitingStatus(globals.GetServer(req.GetToken()), req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup()), nil
}

func (s *GrpcServer) OnRequestShutdown(ctx context.Context, req *protos.RequestMetadataShutdown) (*emptypb.Empty, error) {
	log.Infof("OnRequestShutdown called with token: %s", req.GetToken())
	log.Debugf("Received request metadata: %s %s %d %s %s %v", req.GetMethod(), req.GetRouteParsed(), req.GetStatusCode(), req.GetUser(), req.GetIp(), req.GetApiSpec())

	go storeTotalStats(globals.GetServer(req.GetToken()), req.GetRateLimited())
	go storeRoute(globals.GetServer(req.GetToken()), req.GetMethod(), req.GetRouteParsed(), req.GetApiSpec(), req.GetRateLimited())
	go updateRateLimitingCounts(globals.GetServer(req.GetToken()), req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())

	atomic.StoreUint32(&globals.GetServer(req.GetToken()).GotTraffic, 1)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) GetCloudConfig(ctx context.Context, req *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	log.Infof("GetCloudConfig called with token: %s", req.GetToken())
	cloudConfig := getCloudConfig(globals.GetServer(req.GetToken()), req.GetConfigUpdatedAt())
	if cloudConfig == nil {
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}
	return cloudConfig, nil
}

func (s *GrpcServer) OnUser(ctx context.Context, req *protos.User) (*emptypb.Empty, error) {
	log.Infof("OnUser called with token: %s", req.GetToken())
	log.Debugf("Received user event: %s", req.GetId())
	go onUserEvent(globals.GetServer(req.GetToken()), req.GetId(), req.GetUsername(), req.GetIp())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnAttackDetected(ctx context.Context, req *protos.AttackDetected) (*emptypb.Empty, error) {
	log.Infof("OnAttackDetected called with token: %s", req.GetToken())
	cloud.SendAttackDetectedEvent(globals.GetServer(req.GetToken()), req)
	storeAttackStats(globals.GetServer(req.GetToken()), req)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredSinkStats(ctx context.Context, req *protos.MonitoredSinkStats) (*emptypb.Empty, error) {
	log.Infof("OnMonitoredSinkStats called with token: %s", req.GetToken())
	storeSinkStats(globals.GetServer(req.GetToken()), req)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMiddlewareInstalled(ctx context.Context, req *protos.MiddlewareInstalledInfo) (*emptypb.Empty, error) {
	log.Infof("OnMiddlewareInstalled called with token: %s", req.GetToken())
	log.Debugf("Received MiddlewareInstalled")
	atomic.StoreUint32(&globals.GetServer(req.GetToken()).MiddlewareInstalled, 1)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredIpMatch(ctx context.Context, req *protos.MonitoredIpMatch) (*emptypb.Empty, error) {
	log.Infof("OnMonitoredIpMatch called with token: %s", req.GetToken())
	log.Debugf("Received MonitoredIpMatch: %v", req.GetLists())

	server := globals.GetServer(req.GetToken())
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.IpAddressesMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredUserAgentMatch(ctx context.Context, req *protos.MonitoredUserAgentMatch) (*emptypb.Empty, error) {
	log.Infof("OnMonitoredUserAgentMatch called with token: %s", req.GetToken())
	log.Debugf("Received MonitoredUserAgentMatch: %v", req.GetLists())
	server := globals.GetServer(req.GetToken())
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.UserAgentsMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

var grpcServer *grpc.Server

func StartServer(socketPath string, lis net.Listener) {
	grpcServer = grpc.NewServer() //grpc.MaxConcurrentStreams(100)
	protos.RegisterAikidoServer(grpcServer, &GrpcServer{})

	log.Infof("gRPC server is running on Unix socket %s", socketPath)
	if err := grpcServer.Serve(lis); err != nil {
		log.Warnf("gRPC server failed to serve: %v", err)
	}
	log.Info("gRPC server went down!")
	lis.Close()
}

// Creates the /run/aikido-* folder if it does not exist, in order for the socket creation to succeed
// For now, this folder has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
func createRunDirFolderIfNotExists(socketPath string) {
	runDirectory := filepath.Dir(socketPath)
	if _, err := os.Stat(runDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(runDirectory, 0777)
		if err != nil {
			log.Errorf("Error in creating run directory: %v\n", err)
		} else {
			log.Infof("Run directory %s created successfully.\n", runDirectory)
		}
	} else {
		log.Infof("Run directory %s already exists.\n", runDirectory)
	}
}

func Init(socketPath string) bool {
	// Remove the socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		os.RemoveAll(socketPath)
	}

	createRunDirFolderIfNotExists(socketPath)

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	// Change the permissions of the socket to make it accessible by non-root users
	// For now, this socket has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
	if err := os.Chmod(socketPath, 0777); err != nil {
		panic(fmt.Sprintf("failed to change permissions of Unix socket: %v", err))
	}

	go StartServer(socketPath, lis)
	return true
}

func Uninit() {
	if grpcServer != nil {
		grpcServer.Stop()
		log.Infof("gRPC server has been stopped!")
	}

	// Remove the socket file if it exists
	if _, err := os.Stat(globals.EnvironmentConfig.SocketPath); err == nil {
		if err := os.RemoveAll(globals.EnvironmentConfig.SocketPath); err != nil {
			panic(fmt.Sprintf("failed to remove existing socket: %v", err))
		}
	}
}
