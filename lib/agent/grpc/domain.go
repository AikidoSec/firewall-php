package grpc

import (
	"main/globals"
	"main/log"
)

func storeDomain(domain string, port uint32) {
	if port == 0 {
		return
	}

	globals.HostnamesMutex.Lock()
	defer globals.HostnamesMutex.Unlock()

	if len(globals.Hostnames) >= globals.MaxNumberOfStoredHostnames {
		log.Warnf("Max number of stored hostnames reached, skipping domain %s", domain)
		return
	}

	if _, ok := globals.Hostnames[domain]; !ok {
		globals.Hostnames[domain] = make(map[uint32]uint64)
	}

	globals.Hostnames[domain][port]++
}
