package grpc

import (
	. "main/aikido_types"
	"main/utils"
)

func storeDomain(server *ServerData, domain string, port uint32) {
	if port == 0 {
		return
	}

	server.HostnamesMutex.Lock()
	defer server.HostnamesMutex.Unlock()

	if _, ok := server.Hostnames[domain]; !ok {
		// First time we see this domain
		server.Hostnames[domain] = make(map[uint32]uint64)
		utils.RemoveOldestFromMapIfMaxExceeded(&server.Hostnames, &server.HostnamesQueue, domain)
	}

	server.Hostnames[domain][port]++
}
