package cloud

import (
	. "main/aikido_types"
	"main/utils"
	"sync/atomic"
)

func GetHostnamesAndClear(server *ServerData) []Hostname {
	server.HostnamesMutex.Lock()
	defer server.HostnamesMutex.Unlock()

	var hostnames []Hostname
	for domain := range server.Hostnames {
		for port := range server.Hostnames[domain] {
			hostnames = append(hostnames, Hostname{URL: domain, Port: port, Hits: server.Hostnames[domain][port]})
		}
	}

	server.Hostnames = make(map[string]map[uint32]uint64)
	server.HostnamesQueue.Clear()
	return hostnames
}

func GetRoutesAndClear(server *ServerData) []Route {
	server.RoutesMutex.Lock()
	defer server.RoutesMutex.Unlock()

	var routes []Route
	for _, methodsMap := range server.Routes {
		for _, routeData := range methodsMap {
			if routeData.Hits == 0 {
				continue
			}
			routes = append(routes, *routeData)
			routeData.Hits = 0
		}
	}

	// Clear routes data
	server.Routes = make(map[string]map[string]*Route)
	server.RoutesQueue.Clear()
	return routes
}

func GetUsersAndClear(server *ServerData) []User {
	server.UsersMutex.Lock()
	defer server.UsersMutex.Unlock()

	var users []User
	for _, user := range server.Users {
		users = append(users, user)
	}

	server.Users = make(map[string]User)
	server.UsersQueue.Clear()
	return users
}

func GetMonitoredSinkStatsAndClear(server *ServerData) map[string]MonitoredSinkStats {
	monitoredSinkStats := make(map[string]MonitoredSinkStats)
	for sink, stats := range server.StatsData.MonitoredSinkTimings {
		monitoredSinkStats[sink] = MonitoredSinkStats{
			Kind:                  stats.Kind,
			AttacksDetected:       stats.AttacksDetected,
			InterceptorThrewError: stats.InterceptorThrewError,
			WithoutContext:        stats.WithoutContext,
			Total:                 stats.Total,
			CompressedTimings:     stats.CompressedTimings,
		}

		delete(server.StatsData.MonitoredSinkTimings, sink)
	}
	return monitoredSinkStats
}

func GetPackages(server *ServerData) []Package {
	server.PackagesMutex.Lock()
	defer server.PackagesMutex.Unlock()

	packages := []Package{}
	for _, p := range server.Packages {
		packages = append(packages, p)
	}

	return packages
}

func GetIpsBreakdownAndClear(server *ServerData) MonitoredListsBreakdown {
	m := MonitoredListsBreakdown{
		Breakdown: server.StatsData.IpAddressesMatches,
	}
	server.StatsData.IpAddressesMatches = make(map[string]int)
	return m
}

func GetUserAgentsBreakdownAndClear(server *ServerData) MonitoredListsBreakdown {
	m := MonitoredListsBreakdown{
		Breakdown: server.StatsData.UserAgentsMatches,
	}
	server.StatsData.UserAgentsMatches = make(map[string]int)
	return m
}

func GetStatsAndClear(server *ServerData) Stats {
	server.StatsData.StatsMutex.Lock()
	defer server.StatsData.StatsMutex.Unlock()

	stats := Stats{
		Operations: GetMonitoredSinkStatsAndClear(server),
		StartedAt:  server.StatsData.StartedAt,
		EndedAt:    utils.GetTime(),
		Requests: Requests{
			Total:       server.StatsData.Requests,
			RateLimited: server.StatsData.RequestsRateLimited,
			Aborted:     server.StatsData.RequestsAborted,
			AttacksDetected: AttacksDetected{
				Total:   server.StatsData.Attacks,
				Blocked: server.StatsData.AttacksBlocked,
			},
		},
		UserAgents:  GetUserAgentsBreakdownAndClear(server),
		IpAddresses: GetIpsBreakdownAndClear(server),
	}

	server.StatsData.StartedAt = utils.GetTime()
	server.StatsData.Requests = 0
	server.StatsData.RequestsAborted = 0
	server.StatsData.Attacks = 0
	server.StatsData.AttacksBlocked = 0

	return stats
}

func GetMiddlewareInstalled(server *ServerData) bool {
	return atomic.LoadUint32(&server.MiddlewareInstalled) == 1
}

func SendHeartbeatEvent(server *ServerData) {
	heartbeatEvent := Heartbeat{
		Type:                "heartbeat",
		Agent:               GetAgentInfo(server),
		Time:                utils.GetTime(),
		Stats:               GetStatsAndClear(server),
		Hostnames:           GetHostnamesAndClear(server),
		Routes:              GetRoutesAndClear(server),
		Users:               GetUsersAndClear(server),
		Packages:            GetPackages(server),
		MiddlewareInstalled: GetMiddlewareInstalled(server),
	}

	response, err := SendCloudRequest(server, server.AikidoConfig.Endpoint, EventsAPI, EventsAPIMethod, heartbeatEvent)
	if err != nil {
		LogCloudRequestError(server, "Error in sending heartbeat event: ", err)
		return
	}
	StoreCloudConfig(server, response)
}
