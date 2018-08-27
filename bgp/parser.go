package bgp

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output and tries to find bgp sessions with related data
func (c *bgpCollector) Parse(ostype string, output string) ([]BgpSession, error) {
	if ostype != rpc.IOSXE {
		return nil, errors.New("'show bgp all summary' is not implemented for " + ostype)
	}
	items := []BgpSession{}
	neighborRegexp, _ := regexp.Compile(`^(\S+)\s+\d\s+(\d+)\s+(\d+)\s+(\d+)\s+\d+\s+\d+\s+\d+\s+\S+\s+(\S+)\s*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := neighborRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		pref := util.Str2float64(matches[5])
		up := true
		if pref < 0 {
			pref = 0
			up = false
		}

		item := BgpSession{
			IP:               matches[1],
			Asn:              matches[2],
			InputMessages:    util.Str2float64(matches[3]),
			OutputMessages:   util.Str2float64(matches[4]),
			Up:               up,
			ReceivedPrefixes: pref,
		}
		items = append(items, item)
	}
	return items, nil
}
