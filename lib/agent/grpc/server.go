package grpc

import (
	"context"
	"fmt"
	"main/cloud"
	"main/config"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	protos.AikidoServer
}

func (s *server) OnConfig(ctx context.Context, req *protos.Config) (*emptypb.Empty, error) {
	previousToken := config.GetToken()
	if previousToken == "" {
		storeConfig(req.GetToken(), req.GetLogLevel(), req.GetBlocking(), req.GetLocalhostAllowedByDefault(), req.GetCollectApiSchema())

		// First time the token is set -> we can start reporting things to cloud
		cloud.SendStartEvent()
	}
	return &emptypb.Empty{}, nil
}

func (s *server) OnDomain(ctx context.Context, req *protos.Domain) (*emptypb.Empty, error) {
	log.Debugf("Received domain: %s:%d", req.GetDomain(), req.GetPort())
	storeDomain(req.GetDomain(), req.GetPort())
	return &emptypb.Empty{}, nil
}

func (s *server) GetRateLimitingStatus(ctx context.Context, req *protos.RateLimitingInfo) (*protos.RateLimitingStatus, error) {
	log.Debugf("Received rate limiting info: %s %s %s %s", req.GetMethod(), req.GetRoute(), req.GetUser(), req.GetIp())

	return getRateLimitingStatus(req.GetMethod(), req.GetRoute(), req.GetUser(), req.GetIp()), nil
}

func (s *server) OnRequestShutdown(ctx context.Context, req *protos.RequestMetadataShutdown) (*emptypb.Empty, error) {
	log.Debugf("Received request metadata: %s %s %d %s %s %v", req.GetMethod(), req.GetRoute(), req.GetStatusCode(), req.GetUser(), req.GetIp(), req.GetApiSpec())

	go storeStats()
	go storeRoute(req.GetMethod(), req.GetRoute(), req.GetApiSpec())
	go updateRateLimitingCounts(req.GetMethod(), req.GetRoute(), req.GetUser(), req.GetIp())

	return &emptypb.Empty{}, nil
}

func (s *server) GetCloudConfig(ctx context.Context, req *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	cloudConfig := getCloudConfig(req.ConfigUpdatedAt)
	if cloudConfig == nil {
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}
	return cloudConfig, nil
}

func (s *server) OnUser(ctx context.Context, req *protos.User) (*emptypb.Empty, error) {
	log.Debugf("Received user event: %s", req.GetId())
	go onUserEvent(req.GetId(), req.GetUsername(), req.GetIp())
	return &emptypb.Empty{}, nil
}

func (s *server) OnAttackDetected(ctx context.Context, req *protos.AttackDetected) (*emptypb.Empty, error) {
	cloud.SendAttackDetectedEvent(req)
	storeAttackStats(req)
	return &emptypb.Empty{}, nil
}

func (s *server) OnMonitoredSinkStats(ctx context.Context, req *protos.MonitoredSinkStats) (*emptypb.Empty, error) {
	storeSinkStats(req)
	return &emptypb.Empty{}, nil
}

func (s *server) OnMiddlewareInstalled(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log.Debugf("Received MiddlewareInstalled")
	atomic.StoreUint32(&globals.MiddlewareInstalled, 1)
	return &emptypb.Empty{}, nil
}

var grpcServer *grpc.Server

func StartServer(lis net.Listener) {
	grpcServer = grpc.NewServer()
	protos.RegisterAikidoServer(grpcServer, &server{})

	log.Infof("Server is running on Unix socket %s", globals.EnvironmentConfig.SocketPath)
	if err := grpcServer.Serve(lis); err != nil {
		log.Warnf("gRPC server failed to serve: %v", err)
	}
	log.Info("gRPC server went down!")
	lis.Close()
}

// Creates the /run/aikido-* folder if it does not exist, in order for the socket creation to succeed
// For now, this folder has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
func createRunDirFolderIfNotExists() {
	runDirectory := filepath.Dir(globals.EnvironmentConfig.SocketPath)
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

func Init() bool {
	// Remove the socket file if it already exists
	if _, err := os.Stat(globals.EnvironmentConfig.SocketPath); err == nil {
		os.RemoveAll(globals.EnvironmentConfig.SocketPath)
	}

	createRunDirFolderIfNotExists()

	lis, err := net.Listen("unix", globals.EnvironmentConfig.SocketPath)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	// Change the permissions of the socket to make it accessible by non-root users
	// For now, this socket has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
	if err := os.Chmod(globals.EnvironmentConfig.SocketPath, 0777); err != nil {
		panic(fmt.Sprintf("failed to change permissions of Unix socket: %v", err))
	}

	go StartServer(lis)
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
