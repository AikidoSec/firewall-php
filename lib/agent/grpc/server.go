package grpc

import (
	"context"
	"fmt"
	attackwavedetection "main/attack-wave-detection"
	"main/cloud"
	"main/constants"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"main/server_utils"
	"main/utils"
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

func (s *GrpcServer) OnConfig(ctx context.Context, req *protos.Config) (*emptypb.Empty, error) {
	token := req.GetToken()
	if token == "" {
		return &emptypb.Empty{}, nil
	}

	server := globals.GetServer(token)
	if server != nil {
		log.Debugf(server.Logger, "Server %s already exists, skipping config update...", token)
		return &emptypb.Empty{}, nil
	}

	server_utils.Register(token, req)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnPackages(ctx context.Context, req *protos.Packages) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	storePackages(server, req.GetPackages())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnDomain(ctx context.Context, req *protos.Domain) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received domain: %s:%d", req.GetDomain(), req.GetPort())
	storeDomain(server, req.GetDomain(), req.GetPort())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) GetRateLimitingStatus(ctx context.Context, req *protos.RateLimitingInfo) (*protos.RateLimitingStatus, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &protos.RateLimitingStatus{Block: false}, nil
	}
	log.Debugf(server.Logger, "Received rate limiting info: %s %s %s %s %s %s", req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())
	return getRateLimitingStatus(server, req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup()), nil
}

func (s *GrpcServer) OnRequestShutdown(ctx context.Context, req *protos.RequestMetadataShutdown) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received request metadata: %s %s %d %s %s %v", req.GetMethod(), req.GetRouteParsed(), req.GetStatusCode(), req.GetUser(), req.GetIp(), req.GetApiSpec())
	if req.GetShouldDiscoverRoute() || req.GetRateLimited() {
		go storeTotalStats(server, req.GetRateLimited())
		go storeRoute(server, req.GetMethod(), req.GetRouteParsed(), req.GetApiSpec(), req.GetRateLimited())
		go updateRateLimitingCounts(server, req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())
	}
	if req.GetIsWebScanner() {
		go attackwavedetection.IncrementAndDetect(server, req.GetIp())
	}

	atomic.StoreUint32(&server.GotTraffic, 1)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) GetCloudConfig(ctx context.Context, req *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		log.Warnf(log.MainLogger, "Server \"AIK_RUNTIME_***%s\" not found, returning nil", utils.AnonymizeToken(req.GetToken()))
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}

	atomic.StoreInt64(&server.LastConnectionTime, utils.GetTime())
	cloudConfig := getCloudConfig(server, req.GetConfigUpdatedAt())
	if cloudConfig == nil {
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}
	log.Debugf(server.Logger, "Returning cloud config update to request processor!")
	return cloudConfig, nil
}

func (s *GrpcServer) OnUser(ctx context.Context, req *protos.User) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received user event: %s", req.GetId())
	go onUserEvent(server, req.GetId(), req.GetUsername(), req.GetIp())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnAttackDetected(ctx context.Context, req *protos.AttackDetected) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	cloud.SendAttackDetectedEvent(server, req, "detected_attack")
	storeAttackStats(server, req)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredSinkStats(ctx context.Context, req *protos.MonitoredSinkStats) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	storeSinkStats(server, req)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMiddlewareInstalled(ctx context.Context, req *protos.MiddlewareInstalledInfo) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MiddlewareInstalled")
	atomic.StoreUint32(&server.MiddlewareInstalled, 1)
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredIpMatch(ctx context.Context, req *protos.MonitoredIpMatch) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MonitoredIpMatch: %v", req.GetLists())

	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.IpAddressesMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredUserAgentMatch(ctx context.Context, req *protos.MonitoredUserAgentMatch) (*emptypb.Empty, error) {
	server := globals.GetServer(req.GetToken())
	if server == nil {
		return &emptypb.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MonitoredUserAgentMatch: %v", req.GetLists())

	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.UserAgentsMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

var grpcServer *grpc.Server

func StartServer(lis net.Listener) {
	grpcServer = grpc.NewServer() //grpc.MaxConcurrentStreams(100)
	protos.RegisterAikidoServer(grpcServer, &GrpcServer{})

	log.Infof(log.MainLogger, "gRPC server is running on Unix socket %s", constants.SocketPath)
	if err := grpcServer.Serve(lis); err != nil {
		log.Warnf(log.MainLogger, "gRPC server failed to serve: %v", err)
	}
	log.Info(log.MainLogger, "gRPC server went down!")
	lis.Close()
}

// Creates the /run/aikido-* folder if it does not exist, in order for the socket creation to succeed
// For now, this folder has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
func createRunDirFolderIfNotExists() {
	runDirectory := filepath.Dir(constants.SocketPath)
	if _, err := os.Stat(runDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(runDirectory, 0777)
		if err != nil {
			log.Errorf(log.MainLogger, "Error in creating run directory: %v\n", err)
		} else {
			log.Infof(log.MainLogger, "Run directory %s created successfully.\n", runDirectory)
		}
	} else {
		log.Infof(log.MainLogger, "Run directory %s already exists.\n", runDirectory)
	}
}

func Init() bool {
	// Remove the socket file if it already exists
	if _, err := os.Stat(constants.SocketPath); err == nil {
		os.RemoveAll(constants.SocketPath)
	}

	createRunDirFolderIfNotExists()

	lis, err := net.Listen("unix", constants.SocketPath)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	// Change the permissions of the socket to make it accessible by non-root users
	// For now, this socket has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
	if err := os.Chmod(constants.SocketPath, 0777); err != nil {
		panic(fmt.Sprintf("failed to change permissions of Unix socket: %v", err))
	}

	go StartServer(lis)
	return true
}

func Uninit() {
	if grpcServer != nil {
		grpcServer.Stop()
		log.Infof(log.MainLogger, "gRPC server has been stopped!")
	}

	// Remove the socket file if it exists
	if _, err := os.Stat(constants.SocketPath); err == nil {
		if err := os.RemoveAll(constants.SocketPath); err != nil {
			panic(fmt.Sprintf("failed to remove existing socket: %v", err))
		}
	}
}
