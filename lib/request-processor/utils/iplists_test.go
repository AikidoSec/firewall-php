package utils

import (
	. "main/aikido_types"
	"main/globals"
	"testing"
)

func TestIsIpBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.0.0/16"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	for _, ip := range []string{"1.2.3.4", "::ffff:1.2.3.4", "::ffff:102:304"} {
		t.Run(ip, func(t *testing.T) {
			result, _ := IsIpBlocked(nil, server, ip)
			if result != true {
				t.Errorf("expected true, got %v", result)
			}
		})
	}
}

func TestIsIpBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.3.4"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	for _, ip := range []string{"1.2.3.4", "::ffff:1.2.3.4", "::ffff:102:304"} {
		t.Run(ip, func(t *testing.T) {
			result, _ := IsIpBlocked(nil, server, ip)
			if result != true {
				t.Errorf("expected true, got %v", result)
			}
		})
	}
}

func TestMappedIpMatchesConfiguredEntries(t *testing.T) {
	for _, entry := range []string{"1.2.3.4", "1.2.0.0/16", "::ffff:1.2.3.4", "::ffff:1.2.0.0/112", "::/0"} {
		t.Run(entry, func(t *testing.T) {
			list, err := BuildIpList("test", []string{entry})
			if err != nil {
				t.Fatal(err)
			}
			server := globals.NewServerData()
			server.CloudConfig.BlockedIps = map[string]IpList{"test": *list}
			for _, ip := range []string{"1.2.3.4", "::ffff:1.2.3.4", "::ffff:102:304"} {
				result, _ := IsIpBlocked(nil, server, ip)
				if !result {
					t.Errorf("expected %s to match %s", ip, entry)
				}
				if result := IsIpInSet(nil, &list.IpSet, ip); result != Found {
					t.Errorf("expected IsIpInSet to match %s against %s, got %d", ip, entry, result)
				}
			}
		})
	}
}

func TestIsIpNotBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.0.0/16"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	for _, ip := range []string{"2.3.4.5", "::ffff:2.3.4.5"} {
		t.Run(ip, func(t *testing.T) {
			result, _ := IsIpBlocked(nil, server, ip)
			if result != false {
				t.Errorf("expected false, got %v", result)
			}
		})
	}
}

func TestIsIpNotBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.3.4"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	for _, ip := range []string{"2.3.4.5", "::ffff:2.3.4.5"} {
		t.Run(ip, func(t *testing.T) {
			result, _ := IsIpBlocked(nil, server, ip)
			if result != false {
				t.Errorf("expected false, got %v", result)
			}
		})
	}
}
func TestIsIpv6BlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::/32"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8:1234:5678:90ab:cdef:1234:5678"
	result, _ := IsIpBlocked(nil, server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpv6BlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::1"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8::1"
	result, _ := IsIpBlocked(nil, server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpv6NotBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::/32"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db9::1"
	result, _ := IsIpBlocked(nil, server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestIsIpv6NotBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::1"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8::2"
	result, _ := IsIpBlocked(nil, server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}
