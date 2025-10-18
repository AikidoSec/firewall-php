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
	OnPackages(globals.AikidoConfig.Packages)
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

	_, err := client.OnConfig(ctx, &protos.Config{
		PlatformName:              globals.AikidoConfig.PlatformName,
		PlatformVersion:           globals.AikidoConfig.PlatformVersion,
		Token:                     globals.AikidoConfig.Token,
		Endpoint:                  globals.AikidoConfig.Endpoint,
		ConfigEndpoint:            globals.AikidoConfig.ConfigEndpoint,
		LogLevel:                  globals.AikidoConfig.LogLevel,
		DiskLogs:                  globals.AikidoConfig.DiskLogs,
		Blocking:                  globals.AikidoConfig.Blocking,
		LocalhostAllowedByDefault: globals.AikidoConfig.LocalhostAllowedByDefault,
		CollectApiSchema:          globals.AikidoConfig.CollectApiSchema})
	if err != nil {
		log.Warnf("Could not send Aikido Config: %v", err)
		return
	}

	log.Debugf("Aikido config sent via socket: %+v", globals.AikidoConfig)
}

/* Send outgoing domain to Aikido Agent via gRPC */
func OnDomain(domain string, port uint32) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnDomain(ctx, &protos.Domain{Token: globals.AikidoConfig.Token, Domain: domain, Port: port})
	if err != nil {
		log.Warnf("Could not send domain %v: %v", domain, err)
		return
	}

	log.Debugf("Domain sent via socket: %v:%v", domain, port)
}

/* Send packages to Aikido Agent via gRPC */
func OnPackages(packages map[string]string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnPackages(ctx, &protos.Packages{Token: globals.AikidoConfig.Token, Packages: packages})
	if err != nil {
		log.Warnf("Could not send packages %v: %v", packages, err)
		return
	}

	log.Debugf("Packages sent via socket: %v", packages)
}

/* Send request metadata (route & method) to Aikido Agent via gRPC */
func GetRateLimitingStatus(method string, route string, routeParsed string, user string, ip string, rateLimitGroup string, timeout time.Duration) *protos.RateLimitingStatus {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	RateLimitingStatus, err := client.GetRateLimitingStatus(ctx, &protos.RateLimitingInfo{Token: globals.AikidoConfig.Token, Method: method, Route: route, RouteParsed: routeParsed, User: user, Ip: ip, RateLimitGroup: rateLimitGroup})
	if err != nil {
		log.Warnf("Cannot get rate limiting status %v %v: %v", method, route, err)
		return nil
	}

	log.Debugf("Rate limiting status for (%v %v) sent via socket and got reply (%v)", method, route, RateLimitingStatus)
	return RateLimitingStatus
}

/* Send request metadata (route, method & status code) to Aikido Agent via gRPC */
func OnRequestShutdown(method string, route string, routeParsed string, statusCode int, user string, ip string, rateLimitGroup string, apiSpec *protos.APISpec, rateLimited bool) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnRequestShutdown(ctx, &protos.RequestMetadataShutdown{Token: globals.AikidoConfig.Token, Method: method, Route: route, RouteParsed: routeParsed, StatusCode: int32(statusCode), User: user, Ip: ip, RateLimitGroup: rateLimitGroup, ApiSpec: apiSpec, RateLimited: rateLimited})
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cloudConfig, err := client.GetCloudConfig(ctx, &protos.CloudConfigUpdatedAt{Token: globals.AikidoConfig.Token, ConfigUpdatedAt: utils.GetCloudConfigUpdatedAt()})
	if err != nil {
		return
	}

	log.Debugf("Got cloud config: %v", cloudConfig)
	setCloudConfig(cloudConfig)
}

func OnUserEvent(id string, username string, ip string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnUser(ctx, &protos.User{Token: globals.AikidoConfig.Token, Id: id, Username: username, Ip: ip})
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnAttackDetected(ctx, attackDetected)
	if err != nil {
		log.Warnf("Could not send attack detected event")
		return
	}
	log.Debugf("Attack detected event sent via socket")
}

func OnMonitoredSinkStats(sink, kind string, attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total int32, timings []int64) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnMonitoredSinkStats(ctx, &protos.MonitoredSinkStats{
		Token:                 globals.AikidoConfig.Token,
		Sink:                  sink,
		Kind:                  kind,
		AttacksDetected:       attacksDetected,
		AttacksBlocked:        attacksBlocked,
		InterceptorThrewError: interceptorThrewError,
		WithoutContext:        withoutContext,
		Total:                 total,
		Timings:               timings,
	})
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnMiddlewareInstalled(ctx, &protos.MiddlewareInstalledInfo{Token: globals.AikidoConfig.Token})
	if err != nil {
		log.Warnf("Could not call OnMiddlewareInstalled")
		return
	}
	log.Debugf("OnMiddlewareInstalled sent via socket")
}

func OnMonitoredIpMatch(lists []utils.IpListMatch) {
	if client == nil || len(lists) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	protosLists := []string{}
	for _, list := range lists {
		protosLists = append(protosLists, list.Key)
	}

	_, err := client.OnMonitoredIpMatch(ctx, &protos.MonitoredIpMatch{Token: globals.AikidoConfig.Token, Lists: protosLists})
	if err != nil {
		log.Warnf("Could not call OnMonitoredIpMatch")
		return
	}
	log.Debugf("OnMonitoredIpMatch sent via socket")
}

func OnMonitoredUserAgentMatch(lists []string) {
	if client == nil || len(lists) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnMonitoredUserAgentMatch(ctx, &protos.MonitoredUserAgentMatch{Token: globals.AikidoConfig.Token, Lists: lists})
	if err != nil {
		log.Warnf("Could not call OnMonitoredUserAgentMatch")
		return
	}
	log.Debugf("OnMonitoredUserAgentMatch sent via socket")
}
