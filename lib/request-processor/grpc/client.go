package grpc

import (
	"context"
	"fmt"
	"main/globals"
	"main/log"
	"main/utils"
	"time"

	"main/ipc/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var conn *grpc.ClientConn
var client protos.AikidoClient

func Init() {
	conn, err := grpc.Dial(
		"unix://"+globals.EnvironmentConfig.SocketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		panic(fmt.Sprintf("did not connect: %v", err))
	}

	client = protos.NewAikidoClient(conn)

	log.Debugf("Current connection state: %s\n", conn.GetState().String())

	SendAikidoConfig()
	startCloudConfigRoutine()
}

func Uninit() {
	stopCloudConfigRoutine()
	if conn != nil {
		conn.Close()
	}
}

/* Send Aikido Config to Aikido Agent via gRPC */
func SendAikidoConfig() {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnConfig(ctx, &protos.Config{Token: globals.AikidoConfig.Token, LogLevel: globals.AikidoConfig.LogLevel,
		Blocking: globals.AikidoConfig.Blocking, LocalhostAllowedByDefault: globals.AikidoConfig.LocalhostAllowedByDefault,
		CollectApiSchema: globals.AikidoConfig.CollectApiSchema})
	if err != nil {
		log.Warnf("Could not send Aikido Config: %v", err)
		return
	}

	log.Debugf("Aikido config sent via socket!")
}

/* Send outgoing domain to Aikido Agent via gRPC */
func OnDomain(domain string, port int) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnDomain(ctx, &protos.Domain{Domain: domain, Port: int32(port)})
	if err != nil {
		log.Warnf("Could not send domain %v: %v", domain, err)
		return
	}

	log.Debugf("Domain sent via socket: %v:%v", domain, port)
}

/* Send request metadata (route & method) to Aikido Agent via gRPC */
func GetRateLimitingStatus(method string, route string, user string, ip string, timeout time.Duration) *protos.RateLimitingStatus {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	RateLimitingStatus, err := client.GetRateLimitingStatus(ctx, &protos.RateLimitingInfo{Method: method, Route: route, User: user, Ip: ip})
	if err != nil {
		log.Warnf("Cannot get rate limiting status %v %v: %v", method, route, err)
		return nil
	}

	log.Debugf("Rate limiting status for (%v %v) sent via socket and got reply (%v)", method, route, RateLimitingStatus)
	return RateLimitingStatus
}

/* Send request metadata (route, method & status code) to Aikido Agent via gRPC */
func OnRequestShutdown(method string, route string, statusCode int, user string, ip string, apiSpec *protos.APISpec) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.OnRequestShutdown(ctx, &protos.RequestMetadataShutdown{Method: method, Route: route, StatusCode: int32(statusCode), User: user, Ip: ip, ApiSpec: apiSpec})
	if err != nil {
		log.Warnf("Could not send request metadata %v %v %v: %v", method, route, statusCode, err)
		return
	}

	log.Debugf("Request metadata sent via socket (%v %v %v)", method, route, statusCode)
}

/* Get latest cloud config from Aikido Agent via gRPC */
func GetCloudConfig() {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cloudConfig, err := client.GetCloudConfig(ctx, &protos.CloudConfigUpdatedAt{ConfigUpdatedAt: utils.GetCloudConfigUpdatedAt()})
	if err != nil {
		log.Infof("Could not get cloud config: %v", err)
		return
	}

	log.Debugf("Got cloud config: %v", cloudConfig)
	setCloudConfig(cloudConfig)
}

func OnUserEvent(id string, username string, ip string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnUser(ctx, &protos.User{Id: id, Username: username, Ip: ip})
	if err != nil {
		log.Warnf("Could not send user event %v %v %v: %v", id, username, ip, err)
		return
	}

	log.Debugf("User event sent via socket (%v %v %v)", id, username, ip)
}

func OnAttackDetected(attackDetected *protos.AttackDetected) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnAttackDetected(ctx, attackDetected)
	if err != nil {
		log.Warnf("Could not send attack detected event")
		return
	}
	log.Debugf("Attack detected event sent via socket")
}

func OnMonitoredSinkStats(sink string, attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total int32, averageInMs float64, percentiles map[string]float64) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnMonitoredSinkStats(ctx, &protos.MonitoredSinkStats{Sink: sink, AttacksDetected: attacksDetected,
		InterceptorThrewError: interceptorThrewError, WithoutContext: withoutContext, Total: total,
		CompressedTiming: &protos.CompressedTiming{
			AverageInMs: averageInMs,
			Percentiles: percentiles,
		}})
	if err != nil {
		log.Warnf("Could not send monitored sink stats event")
		return
	}
	log.Debugf("Monitored sink stats for sink \"%s\" sent via socket", sink)
}

func OnMiddlewareInstalled() {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnMiddlewareInstalled(ctx, &emptypb.Empty{})
	if err != nil {
		log.Warnf("Could not call OnMiddlewareInstalled")
		return
	}
	log.Debugf("OnMiddlewareInstalled sent via socket")
}
