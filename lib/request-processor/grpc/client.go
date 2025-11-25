package grpc

import (
	"context"
	"fmt"
	"main/globals"
	"main/log"
	"main/utils"
	"time"

	. "main/aikido_types"

	"main/ipc/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conn *grpc.ClientConn
var client protos.AikidoClient

func Init() {
	conn, err := grpc.Dial(
		"unix://"+globals.SocketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		panic(fmt.Sprintf("did not connect: %v", err))
	}

	client = protos.NewAikidoClient(conn)

	log.Debugf("Current connection state: %s\n", conn.GetState().String())
}

func Uninit() {
	stopCloudConfigRoutine()
	if conn != nil {
		conn.Close()
	}
}

/* Send Aikido Config to Aikido Agent via gRPC */
func SendAikidoConfig(server *ServerData) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnConfig(ctx, &protos.Config{
		Token:                     server.AikidoConfig.Token,
		ServerPid:                 globals.EnvironmentConfig.ServerPID,
		PlatformName:              server.AikidoConfig.PlatformName,
		PlatformVersion:           server.AikidoConfig.PlatformVersion,
		Endpoint:                  server.AikidoConfig.Endpoint,
		ConfigEndpoint:            server.AikidoConfig.ConfigEndpoint,
		LogLevel:                  server.AikidoConfig.LogLevel,
		DiskLogs:                  server.AikidoConfig.DiskLogs,
		Blocking:                  server.AikidoConfig.Blocking,
		LocalhostAllowedByDefault: server.AikidoConfig.LocalhostAllowedByDefault,
		CollectApiSchema:          server.AikidoConfig.CollectApiSchema,
		RequestProcessorPid:       globals.EnvironmentConfig.RequestProcessorPID,
	})
	if err != nil {
		log.Warnf("Could not send Aikido Config: %v", err)
		return
	}

	log.Debugf("Aikido config sent via socket!")
}

/* Send outgoing domain to Aikido Agent via gRPC */
func OnDomain(server *ServerData, domain string, port uint32) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnDomain(ctx, &protos.Domain{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Domain: domain, Port: port})
	if err != nil {
		log.Warnf("Could not send domain %v: %v", domain, err)
		return
	}

	log.Debugf("Domain sent via socket: %v:%v", domain, port)
}

/* Send packages to Aikido Agent via gRPC */
func OnPackages(server *ServerData, packages map[string]string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := client.OnPackages(ctx, &protos.Packages{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Packages: packages})
	if err != nil {
		log.Warnf("Could not send packages %v: %v", packages, err)
		return
	}

	log.Debugf("Packages sent via socket!")
}

/* Send request metadata (route & method) to Aikido Agent via gRPC */
func GetRateLimitingStatus(server *ServerData, method string, route string, routeParsed string, user string, ip string, rateLimitGroup string, timeout time.Duration) *protos.RateLimitingStatus {
	if client == nil || server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	RateLimitingStatus, err := client.GetRateLimitingStatus(ctx, &protos.RateLimitingInfo{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Method: method, Route: route, RouteParsed: routeParsed, User: user, Ip: ip, RateLimitGroup: rateLimitGroup})
	if err != nil {
		log.Warnf("Cannot get rate limiting status %v %v: %v", method, route, err)
		return nil
	}

	log.Debugf("Rate limiting status for (%v %v) sent via socket and got reply (%v)", method, route, RateLimitingStatus)
	return RateLimitingStatus
}

/* Send request metadata (route, method & status code) to Aikido Agent via gRPC */
func OnRequestShutdown(params RequestShutdownParams) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnRequestShutdown(ctx, &protos.RequestMetadataShutdown{
		Token:               params.Server.AikidoConfig.Token,
		ServerPid:           globals.EnvironmentConfig.ServerPID,
		Method:              params.Method,
		Route:               params.Route,
		RouteParsed:         params.RouteParsed,
		StatusCode:          int32(params.StatusCode),
		User:                params.User,
		UserAgent:           params.UserAgent,
		Ip:                  params.IP,
		Url:                 params.Url,
		RateLimitGroup:      params.RateLimitGroup,
		ApiSpec:             params.APISpec,
		RateLimited:         params.RateLimited,
		IsWebScanner:        params.IsWebScanner,
		ShouldDiscoverRoute: params.ShouldDiscoverRoute,
	})
	if err != nil {
		log.Warnf("Could not send request metadata %v %v %v: %v", params.Method, params.Route, params.StatusCode, err)
		return
	}

	log.Debugf("Request metadata sent via socket (%v %v %v)", params.Method, params.Route, params.StatusCode)
}

/* Get latest cloud config from Aikido Agent via gRPC */
func GetCloudConfig(server *ServerData, timeout time.Duration) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cloudConfig, err := client.GetCloudConfig(ctx, &protos.CloudConfigUpdatedAt{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, ConfigUpdatedAt: utils.GetCloudConfigUpdatedAt(server)})
	if err != nil {
		log.Debugf("Could not get cloud config for server \"AIK_RUNTIME_***%s\": %v", utils.AnonymizeToken(server.AikidoConfig.Token), err)
		return
	}

	log.Debugf("Got cloud config for server \"AIK_RUNTIME_***%s\"!", utils.AnonymizeToken(server.AikidoConfig.Token))
	setCloudConfig(server, cloudConfig)
}

func GetCloudConfigForAllServers(timeout time.Duration) {
	for _, server := range globals.GetServers() {
		GetCloudConfig(server, timeout)
	}
}

func OnUserEvent(server *ServerData, id string, username string, ip string) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnUser(ctx, &protos.User{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Id: id, Username: username, Ip: ip})
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

func OnMonitoredSinkStats(server *ServerData, sink, kind string, attacksDetected, attacksBlocked, interceptorThrewError, withoutContext, total int32, timings []int64) {
	if client == nil || server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnMonitoredSinkStats(ctx, &protos.MonitoredSinkStats{
		Token:                 server.AikidoConfig.Token,
		ServerPid:             globals.EnvironmentConfig.ServerPID,
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

func OnMiddlewareInstalled(server *ServerData) {
	if client == nil || server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.OnMiddlewareInstalled(ctx, &protos.MiddlewareInstalledInfo{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID})
	if err != nil {
		log.Warnf("Could not call OnMiddlewareInstalled")
		return
	}
	log.Debugf("OnMiddlewareInstalled sent via socket")
}

func OnMonitoredIpMatch(server *ServerData, lists []utils.IpListMatch) {
	if client == nil || len(lists) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	protosLists := []string{}
	for _, list := range lists {
		protosLists = append(protosLists, list.Key)
	}

	_, err := client.OnMonitoredIpMatch(ctx, &protos.MonitoredIpMatch{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Lists: protosLists})
	if err != nil {
		log.Warnf("Could not call OnMonitoredIpMatch")
		return
	}
	log.Debugf("OnMonitoredIpMatch sent via socket")
}

func OnMonitoredUserAgentMatch(server *ServerData, lists []string) {
	if client == nil || len(lists) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.OnMonitoredUserAgentMatch(ctx, &protos.MonitoredUserAgentMatch{Token: server.AikidoConfig.Token, ServerPid: globals.EnvironmentConfig.ServerPID, Lists: lists})
	if err != nil {
		log.Warnf("Could not call OnMonitoredUserAgentMatch")
		return
	}
	log.Debugf("OnMonitoredUserAgentMatch sent via socket")
}
