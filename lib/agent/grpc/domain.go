package grpc

import (
	"main/globals"
	"main/utils"
)

func storeDomain(domain string, port uint32) {
	if port == 0 {
		return
	}

	globals.HostnamesMutex.Lock()
	defer globals.HostnamesMutex.Unlock()

	if _, ok := globals.Hostnames[domain]; !ok {
		// First time we see this domain
		globals.Hostnames[domain] = make(map[uint32]uint64)
		utils.RemoveOldestFromMapIfMaxExceeded(&globals.Hostnames, &globals.HostnamesQueue, domain)
	}

	globals.Hostnames[domain][port]++
}
