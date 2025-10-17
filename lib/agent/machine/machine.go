package machine

import (
	"main/aikido_types"
	"main/globals"
	"main/log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getDomainName() string {
	var domainName string

	cmd := exec.Command("hostname", "--domain")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	domainName = strings.TrimSpace(string(output))
	return domainName
}

func getOSVersion() string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

func Init() {
	globals.Machine = aikido_types.MachineData{
		HostName:   getHostName(),
		DomainName: getDomainName(),
		OS:         runtime.GOOS,
		OSVersion:  getOSVersion(),
		IPAddress:  getIPAddress(),
	}

	log.Infof("Machine info: %+v", globals.Machine)
}
