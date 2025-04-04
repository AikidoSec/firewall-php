package cloud

import (
	. "main/aikido_types"
	"main/globals"
	"main/ipc/protos"
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
	return hostnames
}

func GetFirstType(types []string) string {
	if len(types) == 0 {
		return ""
	}
	return types[0]
}

func GetApiBodyDataSchemaForCore(schema *protos.DataSchema) *ApiBodyDataSchemaForCore {
	if schema == nil {
		return nil
	}
	return &ApiBodyDataSchemaForCore{
		Type:       GetFirstType(schema.Type),
		Properties: schema.Properties,
		Items:      schema.Items,
		Optional:   schema.Optional,
	}
}

func GetApiBodyInfoForCore(body *protos.APIBodyInfo) *ApiBodyInfoForCore {
	if body == nil {
		return nil
	}
	return &ApiBodyInfoForCore{
		Type:   body.Type,
		Schema: GetApiBodyDataSchemaForCore(body.Schema),
	}
}

func GetApiQueryForCore(query *protos.DataSchema) *ApiQueryForCore {
	if query == nil {
		return nil
	}

	return &ApiQueryForCore{
		Type:       GetFirstType(query.Type),
		Properties: query.Properties,
		Items:      query.Items,
		Optional:   query.Optional,
	}
}

func GetApiSpecForCore(apiSpec *protos.APISpec) *ApiSpecForCore {
	if apiSpec == nil {
		return nil
	}
	return &ApiSpecForCore{
		Body:  GetApiBodyInfoForCore(apiSpec.Body),
		Query: GetApiQueryForCore(apiSpec.Query),
		Auth:  apiSpec.Auth,
	}
}

func GetRoutesAndClear() []RouteForCore {
	globals.RoutesMutex.Lock()
	defer globals.RoutesMutex.Unlock()

	var routes []RouteForCore
	for _, methodsMap := range globals.Routes {
		for _, routeData := range methodsMap {
			if routeData.Hits == 0 {
				continue
			}
			routes = append(routes, RouteForCore{
				Path:    routeData.Path,
				Method:  routeData.Method,
				Hits:    routeData.Hits,
				ApiSpec: GetApiSpecForCore(routeData.ApiSpec),
			})
		}
	}

	// Clear routes data
	globals.Routes = make(map[string]map[string]*Route)
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
	return users
}

func GetMonitoredSinkStatsAndClear() map[string]MonitoredSinkStats {
	monitoredSinkStats := make(map[string]MonitoredSinkStats)
	for sink, stats := range globals.StatsData.MonitoredSinkTimings {
		if stats.Total <= globals.MinStatsCollectedForRelevantMetrics {
			continue
		}

		monitoredSinkStats[sink] = MonitoredSinkStats{
			AttacksDetected:       stats.AttacksDetected,
			InterceptorThrewError: stats.InterceptorThrewError,
			WithoutContext:        stats.WithoutContext,
			Total:                 stats.Total,
			CompressedTimings: []CompressedTiming{
				{
					AverageInMS:  utils.ComputeAverage(stats.Timings),
					Percentiles:  utils.ComputePercentiles(stats.Timings),
					CompressedAt: utils.GetTime(),
				},
			},
		}

		delete(globals.StatsData.MonitoredSinkTimings, sink)
	}
	return monitoredSinkStats
}

func GetStatsAndClear() Stats {
	globals.StatsData.StatsMutex.Lock()
	defer globals.StatsData.StatsMutex.Unlock()

	stats := Stats{
		Sinks:     GetMonitoredSinkStatsAndClear(),
		StartedAt: globals.StatsData.StartedAt,
		EndedAt:   utils.GetTime(),
		Requests: Requests{
			Total:   globals.StatsData.Requests,
			Aborted: globals.StatsData.RequestsAborted,
			AttacksDetected: AttacksDetected{
				Total:   globals.StatsData.Attacks,
				Blocked: globals.StatsData.AttacksBlocked,
			},
		},
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
		MiddlewareInstalled: GetMiddlewareInstalled(),
	}

	response, err := SendCloudRequest(globals.EnvironmentConfig.Endpoint, globals.EventsAPI, globals.EventsAPIMethod, heartbeatEvent)
	if err != nil {
		LogCloudRequestError("Error in sending heartbeat event: ", err)
		return
	}
	StoreCloudConfig(response)
}
