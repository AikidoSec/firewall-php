package grpc

import (
	"context"
	"fmt"
	"main/aikido_types"
	"main/cloud"
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
	token := req.GetToken()
	if token == "" {
		return &emptypb.Empty{}, nil
	}

	var server *aikido_types.ServerData
	if globals.Servers[""] != nil {
		globals.Servers[token] = globals.Servers[""]
		delete(globals.Servers, "")
		server = globals.Servers[token]
	} else {
		if _, exists := globals.Servers[token]; exists {
			return &emptypb.Empty{}, status.Errorf(codes.AlreadyExists, "Server with token already exists!")
		}
		server = aikido_types.NewServerData()
		globals.Servers[token] = server
	}

	storeConfig(server, req.GetToken(), req.GetLogLevel(), req.GetDiskLogs(), req.GetBlocking(), req.GetLocalhostAllowedByDefault(), req.GetCollectApiSchema())
	cloud.SendStartEvent(server)
	return &emptypb.Empty{}, nil
}

func (s *server) OnPackages(ctx context.Context, req *protos.Packages) (*emptypb.Empty, error) {
	storePackages(globals.Servers[req.GetToken()], req.GetPackages())
	return &emptypb.Empty{}, nil
}

func (s *server) OnDomain(ctx context.Context, req *protos.Domain) (*emptypb.Empty, error) {
	log.Debugf("Received domain: %s:%d", req.GetDomain(), req.GetPort())
	storeDomain(globals.Servers[req.GetToken()], req.GetDomain(), req.GetPort())
	return &emptypb.Empty{}, nil
}

func (s *server) GetRateLimitingStatus(ctx context.Context, req *protos.RateLimitingInfo) (*protos.RateLimitingStatus, error) {
	log.Debugf("Received rate limiting info: %s %s %s %s %s %s", req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())

	return getRateLimitingStatus(globals.Servers[req.GetToken()], req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup()), nil
}

func (s *server) OnRequestShutdown(ctx context.Context, req *protos.RequestMetadataShutdown) (*emptypb.Empty, error) {
	log.Debugf("Received request metadata: %s %s %d %s %s %v", req.GetMethod(), req.GetRouteParsed(), req.GetStatusCode(), req.GetUser(), req.GetIp(), req.GetApiSpec())

	go storeTotalStats(globals.Servers[req.GetToken()], req.GetRateLimited())
	go storeRoute(globals.Servers[req.GetToken()], req.GetMethod(), req.GetRouteParsed(), req.GetApiSpec(), req.GetRateLimited())
	go updateRateLimitingCounts(globals.Servers[req.GetToken()], req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())

	atomic.StoreUint32(&globals.Servers[req.GetToken()].GotTraffic, 1)
	return &emptypb.Empty{}, nil
}

func (s *server) GetCloudConfig(ctx context.Context, req *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	cloudConfig := getCloudConfig(globals.Servers[req.GetToken()], req.GetConfigUpdatedAt())
	if cloudConfig == nil {
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}
	return cloudConfig, nil
}

func (s *server) OnUser(ctx context.Context, req *protos.User) (*emptypb.Empty, error) {
	log.Debugf("Received user event: %s", req.GetId())
	go onUserEvent(globals.Servers[req.GetToken()], req.GetId(), req.GetUsername(), req.GetIp())
	return &emptypb.Empty{}, nil
}

func (s *server) OnAttackDetected(ctx context.Context, req *protos.AttackDetected) (*emptypb.Empty, error) {
	cloud.SendAttackDetectedEvent(globals.Servers[req.GetToken()], req)
	storeAttackStats(globals.Servers[req.GetToken()], req)
	return &emptypb.Empty{}, nil
}

func (s *server) OnMonitoredSinkStats(ctx context.Context, req *protos.MonitoredSinkStats) (*emptypb.Empty, error) {
	storeSinkStats(globals.Servers[req.GetToken()], req)
	return &emptypb.Empty{}, nil
}

func (s *server) OnMiddlewareInstalled(ctx context.Context, req *protos.MiddlewareInstalledInfo) (*emptypb.Empty, error) {
	log.Debugf("Received MiddlewareInstalled")
	atomic.StoreUint32(&globals.Servers[req.GetToken()].MiddlewareInstalled, 1)
	return &emptypb.Empty{}, nil
}

func (s *server) OnMonitoredIpMatch(ctx context.Context, req *protos.MonitoredIpMatch) (*emptypb.Empty, error) {
	log.Debugf("Received MonitoredIpMatch: %v", req.GetLists())
	globals.Servers[req.GetToken()].StatsData.StatsMutex.Lock()
	defer globals.Servers[req.GetToken()].StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&globals.Servers[req.GetToken()].StatsData.IpAddressesMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

func (s *server) OnMonitoredUserAgentMatch(ctx context.Context, req *protos.MonitoredUserAgentMatch) (*emptypb.Empty, error) {
	log.Debugf("Received MonitoredUserAgentMatch: %v", req.GetLists())
	globals.Servers[req.GetToken()].StatsData.StatsMutex.Lock()
	defer globals.Servers[req.GetToken()].StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&globals.Servers[req.GetToken()].StatsData.UserAgentsMatches, req.GetLists())
	return &emptypb.Empty{}, nil
}

var grpcServer *grpc.Server

func StartServer(serverObject *aikido_types.ServerData, lis net.Listener) {
	grpcServer = grpc.NewServer()
	protos.RegisterAikidoServer(grpcServer, &server{})

	log.Infof("Server is running on Unix socket %s", serverObject.EnvironmentConfig.SocketPath)
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

func Init(serverObject *aikido_types.ServerData) bool {
	// Remove the socket file if it already exists
	if _, err := os.Stat(serverObject.EnvironmentConfig.SocketPath); err == nil {
		os.RemoveAll(serverObject.EnvironmentConfig.SocketPath)
	}

	createRunDirFolderIfNotExists(serverObject.EnvironmentConfig.SocketPath)

	lis, err := net.Listen("unix", serverObject.EnvironmentConfig.SocketPath)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	// Change the permissions of the socket to make it accessible by non-root users
	// For now, this socket has 777 permissions as we don't know under which user the php requests will run under (apache, nginx, www-data, forge, ...)
	if err := os.Chmod(serverObject.EnvironmentConfig.SocketPath, 0777); err != nil {
		panic(fmt.Sprintf("failed to change permissions of Unix socket: %v", err))
	}

	go StartServer(serverObject, lis)
	return true
}

func Uninit(serverObject *aikido_types.ServerData) {
	if grpcServer != nil {
		grpcServer.Stop()
		log.Infof("gRPC server has been stopped!")
	}

	// Remove the socket file if it exists
	if _, err := os.Stat(serverObject.EnvironmentConfig.SocketPath); err == nil {
		if err := os.RemoveAll(serverObject.EnvironmentConfig.SocketPath); err != nil {
			panic(fmt.Sprintf("failed to remove existing socket: %v", err))
		}
	}
}
