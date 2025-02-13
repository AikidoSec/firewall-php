package grpc

import (
	"main/globals"
)

func storeDomain(domain string, port uint32) {
	if port == 0 {
		return
	}

	globals.HostnamesMutex.Lock()
	defer globals.HostnamesMutex.Unlock()

	if _, ok := globals.Hostnames[domain]; !ok {
		globals.Hostnames[domain] = make(map[uint32]uint64)
	}

	globals.Hostnames[domain][port]++
}
