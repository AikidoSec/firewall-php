package utils

import (
	"fmt"
	. "main/aikido_types"
	"main/config"
	"main/globals"
	"net"
	"sort"
	"time"
)

func KeyMustExist[K comparable, V any](m map[K]V, key K) {
	if _, exists := m[key]; !exists {
		panic(fmt.Sprintf("Key %v does not exist in map!", key))
	}
}

func GetFromMap[T any](m map[string]interface{}, key string) *T {
	value, ok := m[key]
	if !ok {
		return nil
	}
	result, ok := value.(T)
	if !ok {
		return nil
	}
	return &result
}

func MustGetFromMap[T any](m map[string]interface{}, key string) T {
	value := GetFromMap[T](m, key)
	if value == nil {
		panic(fmt.Sprintf("Error parsing JSON: key %s does not exist or it has an incorrect type", key))
	}
	return *value
}

func ArrayContains(array []string, search string) bool {
	for _, member := range array {
		if member == search {
			return true
		}
	}
	return false
}

func isLocalhost(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsLoopback()
}

func StartPollingRoutine(stopChan chan struct{}, ticker *time.Ticker, pollingFunction func()) {
	go func() {
		for {
			select {
			case <-ticker.C:
				pollingFunction()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func StopPollingRoutine(stopChan chan struct{}) {
	close(stopChan)
}

func IsBlockingEnabled() bool {
	globals.CloudConfigMutex.Lock()
	defer globals.CloudConfigMutex.Unlock()

	if globals.CloudConfig.Block == nil {
		return config.GetBlocking()
	}
	return *globals.CloudConfig.Block
}

func GetTime() int64 {
	return time.Now().UnixMilli()
}

func GetUserById(userId string) *User {
	if userId == "" {
		return nil
	}

	globals.UsersMutex.Lock()
	defer globals.UsersMutex.Unlock()

	user, exists := globals.Users[userId]
	if !exists {
		return nil
	}
	return &user
}

func ComputeAverage(times []int64) float64 {
	if len(times) == 0 {
		return 0
	}
	var total int64
	for _, t := range times {
		total += t
	}

	return float64(total) / float64(len(times)) / 1e6
}

func ComputePercentiles(times []int64) map[string]float64 {
	if len(times) == 0 {
		return map[string]float64{
			"P50": 0,
			"P90": 0,
			"P95": 0,
			"P99": 0,
		}
	}

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	percentiles := map[string]float64{}
	percentiles["P50"] = float64(times[len(times)/2]) / 1e6
	percentiles["P90"] = float64(times[int(0.9*float64(len(times)))]) / 1e6
	percentiles["P95"] = float64(times[int(0.95*float64(len(times)))]) / 1e6
	percentiles["P99"] = float64(times[int(0.99*float64(len(times)))]) / 1e6

	return percentiles
}
