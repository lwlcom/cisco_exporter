package neighbors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/util"
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
func (c *neighborsCollector) ParseIPv4Neighbors(ostype string, output string, data map[string]*InterfaceNeighors) error {
	// example:
	// Dynamic, via Vlan8, last updated 9 minutes ago.
	// Incomplete, via Vlan8, last updated 0 minute ago.
	ipv4NeighborRegexp, _ := regexp.Compile(`^\s*(Dynamic|Incomplete|Interface),? via ([a-zA-Z0-9\/\-\.]+)`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		matches := ipv4NeighborRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		interface_neigbors, ok := data[matches[2]]
		if !ok {
			return errors.New(fmt.Sprintf("Interface %s found in ARP but not 'sh ip int brie' command", matches[2]))
		}
		if matches[1] == "Incomplete" {
			interface_neigbors.Incomplete++
		} else if matches[1] == "Dynamic" {
			interface_neigbors.Reachable++
		}
	}
	return nil
}

// ParseIPv6Neighbors parses cli output and counts neighbor entries per state for an interface
func (c *neighborsCollector) ParseIPv6Neighbors(ostype string, output string, data map[string]*InterfaceNeighors) error {
	// IPv6 Address                              Age Link-layer Addr State Interface
	// FE80::AD79:7159:3AB9:D52F                   0 aaaa.6cd6.0e6f  STALE Vl65
	ipv6NeighborRegexp, _ := regexp.Compile(`^[a-zA-Z0-9\:]+\s+[\d\-]+\s+[a-zA-Z0-9\.]+\s+(\w+)\s+(\S+)`)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		matches := ipv6NeighborRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		iface_name, err := util.InterfaceShortToLong(matches[2])
		if err != nil {
			return err
		}
		interface_neigbors, ok := data[iface_name]
		if !ok {
			return errors.New(fmt.Sprintf("Interface %s found in ipv6 neighbors but not 'show ipv6 interface brief' command", matches[2]))
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
	return nil
}
