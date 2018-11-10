package bgp

import (
	"errors"
	"regexp"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output and tries to find bgp sessions with related data
func (c *bgpCollector) Parse(ostype string, output string) ([]BgpSession, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS {
		return nil, errors.New("'show bgp all summary' is not implemented for " + ostype)
	}
	items := []BgpSession{}
	neighborRegexp, _ := regexp.Compile(`(\S+)\s+\d\s+(\d+)\s+(\d+)\s+(\d+)\s+\d+\s+\d+\s+\d+\s+\S+\s+(\S+)\s*`)

	matches := neighborRegexp.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		pref := util.Str2float64(match[5])
		up := true
		if pref < 0 {
			pref = 0
			up = false
		}

		item := BgpSession{
			IP:               match[1],
			Asn:              match[2],
			InputMessages:    util.Str2float64(match[3]),
			OutputMessages:   util.Str2float64(match[4]),
			Up:               up,
			ReceivedPrefixes: pref,
		}
		items = append(items, item)
	}
	return items, nil
}
