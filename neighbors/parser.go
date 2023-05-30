package neighbors

import (
	"regexp"
	"strings"
)

// ParseInterfacesIPv4 parses cli output and returns list of interface names
func (c *neighborsCollector) ParseInterfacesIPv4(ostype string, output string) ([]string, error) {
	var items []string
	// Interface              IP-Address      OK? Method Status                Protocol
	// Vlan1                  10.66.115.1     YES NVRAM  up                    up
	deviceNameRegexp, _ := regexp.Compile(`^([a-zA-Z0-9\/\.-]+)\s+\d+\.\d+\.\d+\.\d+`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := deviceNameRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		items = append(items, matches[1])
	}
	return items, nil
}

// ParseInterfacesIPv6 parses cli output and returns list of interface names
func (c *neighborsCollector) ParseInterfacesIPv6(ostype string, output string) ([]string, error) {
	var items []string
	deviceNameRegexp, _ := regexp.Compile(`^([a-zA-Z0-9\/\.-]+)\s+`)
	lines := strings.Split(output, "\n")
	var device_name string
	for _, line := range lines {
		matches := deviceNameRegexp.FindStringSubmatch(line)
		if matches == nil {
			if len(device_name) > 0 {
				if !strings.Contains(line, "unassigned") {
					// interface with ipv6 address
					items = append(items, device_name)
				}
			}
			device_name = ""
		} else {
			device_name = matches[1]
		}
	}
	return items, nil
}

// ParseIPv4Neighbors parses cli output and counts neighbor entries per state for an interface
func (c *neighborsCollector) ParseIPv4Neighbors(ostype string, output string) (InterfaceNeighors, error) {
	interface_neigbors := InterfaceNeighors{
		Incomplete: 0,
		Reachable:  0,
		Stale:      0,
		Delay:      0,
		Probe:      0,
	}
	// Protocol  Address          Age (min)  Hardware Addr   Type   Interface
	// Internet  10.172.80.56            -   aaaa.7fc9.aaff  ARPA   TwentyFiveGigE1/0/46
	// Internet  10.172.80.57           10   aaaa.c14e.fbd6  ARPA   TwentyFiveGigE1/0/46
	// Internet  10.172.36.126           0   Incomplete      ARPA
	ipv4NeighborRegexp, _ := regexp.Compile(`^\w+\s+\d+\.\d+\.\d+\.\d+\s+([\d\-]+)\s+([a-zA-Z0-9\.]+)\s+`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		matches := ipv4NeighborRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		if matches[2] == "Incomplete" {
			interface_neigbors.Incomplete++
		} else if matches[1] != "-" {
			interface_neigbors.Reachable++
		}
	}
	return interface_neigbors, nil
}

// ParseIPv6Neighbors parses cli output and counts neighbor entries per state for an interface
func (c *neighborsCollector) ParseIPv6Neighbors(ostype string, output string) (InterfaceNeighors, error) {
	interface_neigbors := InterfaceNeighors{
		Incomplete: 0,
		Reachable:  0,
		Stale:      0,
		Delay:      0,
		Probe:      0,
	}
	// IPv6 Address                              Age Link-layer Addr State Interface
	// FE80::AD79:7159:3AB9:D52F                   0 aaaa.6cd6.0e6f  STALE Vl65
	ipv6NeighborRegexp, _ := regexp.Compile(`^[a-zA-Z0-9\:]+\s+[\d\-]+\s+[a-zA-Z0-9\.]+\s+(\w+)\s+`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		matches := ipv6NeighborRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		if matches[1] == "INCOM" {
			interface_neigbors.Incomplete++
		} else if matches[1] == "REACH" {
			interface_neigbors.Reachable++
		} else if matches[1] == "STALE" {
			interface_neigbors.Stale++
		} else if matches[1] == "DELAY" {
			interface_neigbors.Delay++
		} else if matches[1] == "PROBE" {
			interface_neigbors.Probe++
		}
	}
	return interface_neigbors, nil
}
