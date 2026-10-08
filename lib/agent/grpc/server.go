package grpc

import (
	"context"
	"fmt"
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
	"syscall"

	. "main/aikido_types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	protos.AikidoServer
}

func (s *GrpcServer) OnConfig(ctx context.Context, req *protos.Config) (*protos.Empty, error) {
	token := req.GetToken()
	if token == "" {
		return &protos.Empty{}, nil
	}

	// Validate token format - Aikido tokens should follow a specific pattern
	// This prevents arbitrary token injection by malicious local processes
	if !isValidTokenFormat(token) {
		log.Warnf(log.MainLogger, "Rejected OnConfig: invalid token format from PID %d", getPeerPid(ctx))
		return &protos.Empty{}, status.Error(codes.InvalidArgument, "invalid token format")
	}

	server := globals.GetServer(ServerKey{Token: token, ServerPID: req.GetServerPid()})
	if server != nil {
		log.Debugf(server.Logger, "Server \"AIK_RUNTIME_***%s\" already exists, skipping config update (request processor PID: %d, server PID: %d)", utils.AnonymizeToken(token), req.GetRequestProcessorPid(), req.GetServerPid())
		return &protos.Empty{}, nil
	}

	server_utils.Register(ServerKey{Token: token, ServerPID: req.GetServerPid()}, req.GetRequestProcessorPid(), req)
	return &protos.Empty{}, nil
}

// isValidTokenFormat validates that a token follows the expected Aikido token format
// Aikido tokens are typically alphanumeric strings of a specific length
func isValidTokenFormat(token string) bool {
	// Token should not be empty and should have reasonable length
	if len(token) < 10 || len(token) > 256 {
		return false
	}

	// Token should only contain alphanumeric characters, hyphens, and underscores
	// This prevents injection attacks and ensures only legitimate tokens are accepted
	for _, c := range token {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}

	return true
}

// getPeerPid extracts the peer process ID from the context for logging
func getPeerPid(ctx context.Context) int32 {
	if p, ok := peer.FromContext(ctx); ok {
		if authInfo, ok := p.AuthInfo.(peerAuthInfo); ok {
			return authInfo.Pid
		}
	}
	return -1
}

func (s *GrpcServer) OnPackages(ctx context.Context, req *protos.Packages) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	storePackages(server, req.GetPackages())
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnDomain(ctx context.Context, req *protos.Domain) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received domain: %s:%d", req.GetDomain(), req.GetPort())
	storeDomain(server, req.GetDomain(), req.GetPort())
	return &protos.Empty{}, nil
}

func (s *GrpcServer) GetRateLimitingStatus(ctx context.Context, req *protos.RateLimitingInfo) (*protos.RateLimitingStatus, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.RateLimitingStatus{Block: false}, nil
	}
	log.Debugf(server.Logger, "Received rate limiting info: %s %s %s %s %s %s", req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup())
	return getRateLimitingStatus(server, req.GetMethod(), req.GetRoute(), req.GetRouteParsed(), req.GetUser(), req.GetIp(), req.GetRateLimitGroup()), nil
}

func (s *GrpcServer) OnRequestShutdown(ctx context.Context, req *protos.RequestMetadataShutdown) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received request metadata: %s %s %d %s %s %v", req.GetMethod(), req.GetRouteParsed(), req.GetStatusCode(), req.GetUser(), req.GetIp(), req.GetApiSpec())
	if req.GetShouldDiscoverRoute() || req.GetRateLimited() {
		go storeTotalStats(server, req.GetRateLimited())
		go storeRoute(server, req.GetMethod(), req.GetRouteParsed(), req.GetApiSpec(), req.GetRateLimited())
	}
	go updateAttackWaveCountsAndDetect(server, req.GetIsWebScanner(), req.GetIp(), req.GetUser(), req.GetUserAgent(), req.GetMethod(), req.GetUrl())

	return &protos.Empty{}, nil
}

func (s *GrpcServer) GetCloudConfig(ctx context.Context, req *protos.CloudConfigUpdatedAt) (*protos.CloudConfig, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		log.Warnf(log.MainLogger, "Server \"AIK_RUNTIME_***%s\" not found, returning nil", utils.AnonymizeToken(req.GetToken()))
		return nil, status.Errorf(codes.Canceled, "Server not found")
	}

	atomic.StoreInt64(&server.LastConnectionTime, utils.GetTime())
	cloudConfig := getCloudConfig(server, req.GetConfigUpdatedAt())
	if cloudConfig == nil {
		return nil, status.Errorf(codes.Canceled, "CloudConfig was not updated")
	}
	log.Debugf(server.Logger, "Returning cloud config update for server \"AIK_RUNTIME_***%s\"!", utils.AnonymizeToken(req.GetToken()))
	return cloudConfig, nil
}

func (s *GrpcServer) OnUser(ctx context.Context, req *protos.User) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received user event: %s", req.GetId())
	go onUserEvent(server, req.GetId(), req.GetUsername(), req.GetIp())
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnCustomEvent(ctx context.Context, req *protos.CustomEvent) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}

	cloud.ScheduleCustomEvent(server, req)
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnAttackDetected(ctx context.Context, req *protos.AttackDetected) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	cloud.SendAttackDetectedEvent(server, req, "detected_attack")
	storeAttackStats(server, req)
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredSinkStats(ctx context.Context, req *protos.MonitoredSinkStats) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	storeSinkStats(server, req)
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnMiddlewareInstalled(ctx context.Context, req *protos.MiddlewareInstalledInfo) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MiddlewareInstalled")
	atomic.StoreUint32(&server.MiddlewareInstalled, 1)
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredIpMatch(ctx context.Context, req *protos.MonitoredIpMatch) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MonitoredIpMatch: %v", req.GetLists())

	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.IpAddressesMatches, req.GetLists())
	return &protos.Empty{}, nil
}

func (s *GrpcServer) OnMonitoredUserAgentMatch(ctx context.Context, req *protos.MonitoredUserAgentMatch) (*protos.Empty, error) {
	server := globals.GetServer(ServerKey{Token: req.GetToken(), ServerPID: req.GetServerPid()})
	if server == nil {
		return &protos.Empty{}, nil
	}
	log.Debugf(server.Logger, "Received MonitoredUserAgentMatch: %v", req.GetLists())

	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	storeMonitoredListsMatches(&server.StatsData.UserAgentsMatches, req.GetLists())
	return &protos.Empty{}, nil
}

var grpcServer *grpc.Server

// peerAuthInfo holds Unix socket peer credential information
type peerAuthInfo struct {
	credentials.CommonAuthInfo
	Pid int32
	Uid uint32
	Gid uint32
}

func (p peerAuthInfo) AuthType() string {
	return "unix-peer"
}

// unixPeerAuthenticator validates peer credentials for Unix socket connections
func unixPeerAuthenticator(ctx context.Context) (context.Context, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "no peer information")
	}

	// Extract Unix socket peer credentials
	if authInfo, ok := p.AuthInfo.(peerAuthInfo); ok {
		// Validate that the peer is a legitimate process
		// Check if the process exists and is accessible
		if authInfo.Pid <= 0 {
			log.Warnf(log.MainLogger, "Rejected connection: invalid PID %d", authInfo.Pid)
			return ctx, status.Error(codes.PermissionDenied, "invalid peer process")
		}

		// Verify the process exists by checking /proc/<pid>/exe
		exePath := fmt.Sprintf("/proc/%d/exe", authInfo.Pid)
		if _, err := os.Readlink(exePath); err != nil {
			log.Warnf(log.MainLogger, "Rejected connection: cannot verify process %d: %v", authInfo.Pid, err)
			return ctx, status.Error(codes.PermissionDenied, "cannot verify peer process")
		}

		log.Debugf(log.MainLogger, "Authenticated peer: PID=%d, UID=%d, GID=%d", authInfo.Pid, authInfo.Uid, authInfo.Gid)
		return ctx, nil
	}

	// If we can't get peer credentials, reject the connection
	log.Warnf(log.MainLogger, "Rejected connection: no peer credentials available")
	return ctx, status.Error(codes.Unauthenticated, "peer authentication required")
}

// unaryInterceptor wraps unary RPC calls with peer authentication
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	newCtx, err := unixPeerAuthenticator(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

// streamInterceptor wraps streaming RPC calls with peer authentication
func streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	newCtx, err := unixPeerAuthenticator(ss.Context())
	if err != nil {
		return err
	}

	// Create a wrapped stream with the authenticated context
	wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: newCtx}
	return handler(srv, wrappedStream)
}

// wrappedServerStream wraps a grpc.ServerStream to override the context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// customTransportCredentials implements credentials.TransportCredentials for Unix sockets
type customTransportCredentials struct{}

func (c customTransportCredentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return rawConn, peerAuthInfo{}, nil
}

func (c customTransportCredentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	// Extract peer credentials from Unix socket
	unixConn, ok := rawConn.(*net.UnixConn)
	if !ok {
		return nil, nil, fmt.Errorf("connection is not a Unix socket")
	}

	// Get the underlying file descriptor
	file, err := unixConn.File()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file descriptor: %v", err)
	}
	defer file.Close()

	fd := int(file.Fd())

	// Get peer credentials using SO_PEERCRED
	ucred, err := syscall.GetsockoptUcred(fd, syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get peer credentials: %v", err)
	}

	authInfo := peerAuthInfo{
		Pid: ucred.Pid,
		Uid: ucred.Uid,
		Gid: ucred.Gid,
	}

	return rawConn, authInfo, nil
}

func (c customTransportCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		SecurityProtocol: "unix-peer",
		SecurityVersion:  "1.0",
	}
}

func (c customTransportCredentials) Clone() credentials.TransportCredentials {
	return c
}

func (c customTransportCredentials) OverrideServerName(serverName string) error {
	return nil
}

func StartServer(lis net.Listener) {
	grpcServer = grpc.NewServer(
		grpc.Creds(customTransportCredentials{}),
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB max receive message size
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB max send message size
	)
	protos.RegisterAikidoServer(grpcServer, &GrpcServer{})

	log.Infof(log.MainLogger, "gRPC server is running on Unix socket %s with peer authentication", constants.SocketPath)
	if err := grpcServer.Serve(lis); err != nil {
		log.Warnf(log.MainLogger, "gRPC server failed to serve: %v", err)
	}
	log.Info(log.MainLogger, "gRPC server went down!")
	lis.Close()
}

// Creates the /run/aikido-* folder if it does not exist, in order for the socket creation to succeed
// The folder has 0755 permissions to prevent unauthorized socket replacement while allowing legitimate PHP processes to connect
func createRunDirFolderIfNotExists() {
	runDirectory := filepath.Dir(constants.SocketPath)
	if _, err := os.Stat(runDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(runDirectory, 0755)
		if err != nil {
			log.Errorf(log.MainLogger, "Error in creating run directory: %v\n", err)
		} else {
			log.Infof(log.MainLogger, "Run directory %s created successfully with restricted permissions.\n", runDirectory)
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

	// Change the permissions of the socket to 0770 to allow group access while preventing unauthorized connections
	// This allows legitimate PHP processes running under various users (apache, nginx, www-data, forge, etc.)
	// to connect while preventing arbitrary local processes from accessing the socket
	if err := os.Chmod(constants.SocketPath, 0770); err != nil {
		panic(fmt.Sprintf("failed to change permissions of Unix socket: %v", err))
	}

	log.Infof(log.MainLogger, "Unix socket %s created with restricted permissions (0770)", constants.SocketPath)

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
