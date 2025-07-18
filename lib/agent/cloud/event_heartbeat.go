package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/utils"
	"sync/atomic"
)

func GetHostnamesAndClear() []Hostname {
	globals.HostnamesMutex.Lock()
	defer globals.HostnamesMutex.Unlock()

	var hostnames []Hostname
	for domain := range globals.Hostnames {
		for port := range globals.Hostnames[domain] {
			hostnames = append(hostnames, Hostname{URL: domain, Port: port, Hits: globals.Hostnames[domain][port]})
		}
	}

	globals.Hostnames = make(map[string]map[uint32]uint64)
	globals.HostnamesQueue.Clear()
	return hostnames
}

func GetRoutesAndClear() []Route {
	globals.RoutesMutex.Lock()
	defer globals.RoutesMutex.Unlock()

	var routes []Route
	for _, methodsMap := range globals.Routes {
		for _, routeData := range methodsMap {
			if routeData.Hits == 0 {
				continue
			}
			routes = append(routes, *routeData)
			routeData.Hits = 0
		}
	}

	// Clear routes data
	globals.Routes = make(map[string]map[string]*Route)
	globals.RoutesQueue.Clear()
	return routes
}

func GetUsersAndClear() []User {
	globals.UsersMutex.Lock()
	defer globals.UsersMutex.Unlock()

	var users []User
	for _, user := range globals.Users {
		users = append(users, user)
	}

	globals.Users = make(map[string]User)
	globals.UsersQueue.Clear()
	return users
}

func GetMonitoredSinkStatsAndClear() map[string]MonitoredSinkStats {
	monitoredSinkStats := make(map[string]MonitoredSinkStats)
	for sink, stats := range globals.StatsData.MonitoredSinkTimings {
		monitoredSinkStats[sink] = MonitoredSinkStats{
			Kind:                  stats.Kind,
			AttacksDetected:       stats.AttacksDetected,
			InterceptorThrewError: stats.InterceptorThrewError,
			WithoutContext:        stats.WithoutContext,
			Total:                 stats.Total,
			CompressedTimings:     stats.CompressedTimings,
		}

		delete(globals.StatsData.MonitoredSinkTimings, sink)
	}
	return monitoredSinkStats
}

func GetPackages() []Package {
	globals.PackagesMutex.Lock()
	defer globals.PackagesMutex.Unlock()

	packages := []Package{}
	for _, p := range globals.Packages {
		packages = append(packages, p)
	}

	return packages
}

func GetIpsBreakdownAndClear() MonitoredListsBreakdown {
	m := MonitoredListsBreakdown{
		Breakdown: globals.StatsData.IpAddressesMatches,
	}
	globals.StatsData.IpAddressesMatches = make(map[string]int)
	return m
}

func GetUserAgentsBreakdownAndClear() MonitoredListsBreakdown {
	m := MonitoredListsBreakdown{
		Breakdown: globals.StatsData.UserAgentsMatches,
	}
	globals.StatsData.UserAgentsMatches = make(map[string]int)
	return m
}

func GetStatsAndClear() Stats {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	stats := Stats{
		Operations: GetMonitoredSinkStatsAndClear(),
		StartedAt:  globals.StatsData.StartedAt,
		EndedAt:    utils.GetTime(),
		Requests: Requests{
			Total:       globals.StatsData.Requests,
			RateLimited: globals.StatsData.RequestsRateLimited,
			Aborted:     globals.StatsData.RequestsAborted,
			AttacksDetected: AttacksDetected{
				Total:   globals.StatsData.Attacks,
				Blocked: globals.StatsData.AttacksBlocked,
			},
		},
		UserAgents:  GetUserAgentsBreakdownAndClear(),
		IpAddresses: GetIpsBreakdownAndClear(),
	}

	globals.StatsData.StartedAt = utils.GetTime()
	globals.StatsData.Requests = 0
	globals.StatsData.RequestsAborted = 0
	globals.StatsData.Attacks = 0
	globals.StatsData.AttacksBlocked = 0

	return stats
}

func GetMiddlewareInstalled() bool {
	return atomic.LoadUint32(&globals.MiddlewareInstalled) == 1
}

func SendHeartbeatEvent() {
	heartbeatEvent := Heartbeat{
		Type:                "heartbeat",
		Agent:               GetAgentInfo(),
		Time:                utils.GetTime(),
		Stats:               GetStatsAndClear(),
		Hostnames:           GetHostnamesAndClear(),
		Routes:              GetRoutesAndClear(),
		Users:               GetUsersAndClear(),
		Packages:            GetPackages(),
		MiddlewareInstalled: GetMiddlewareInstalled(),
	}

	response, err := SendCloudRequest(globals.EnvironmentConfig.Endpoint, globals.EventsAPI, globals.EventsAPIMethod, heartbeatEvent)
	if err != nil {
		LogCloudRequestError("Error in sending heartbeat event: ", err)
		return
	}
	StoreCloudConfig(response)
}
