package environment

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output and tries to find oll temperature and power related data
func (c *environmentCollector) Parse(ostype string, output string) ([]EnvironmentItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show environment' is not implemented for " + ostype)
	}
	items := []EnvironmentItem{}
	tempRegexp := make(map[string]*regexp.Regexp)
	powerRegexp := make(map[string]*regexp.Regexp)
	tempRegexp[rpc.IOSXE], _ = regexp.Compile(`\s*Temp: (\w+)\s+(\w+)\s+\w+\s+(\d+) Celsius`)
	powerRegexp[rpc.IOSXE], _ = regexp.Compile(`(PS\d+)\s+([\w\-]+)\s+\w+\s+\d+\s\w+\s+(\w+)`)
	tempRegexp[rpc.IOS], _ = regexp.Compile(`^(\d+)\s+(air \w+(?: +\w+)?)\s+(\d+)C \(.*\)\s+\w+$`)
	powerRegexp[rpc.IOS], _ = regexp.Compile(`^(\w+)\s+.+\s+(AC) \w+\s+(\w+)\s+\w+\s+.+\s+.+$`)
	tempRegexp[rpc.NXOS], _ = regexp.Compile(`^(\d+)\s+(.+)\s+\d\d?\s+\d\d?\s+(\d\d?)\s+\w+\s*$`)
	powerRegexp[rpc.NXOS], _ = regexp.Compile(`^(\d+)\s+.+\s+(AC)\s+.+\s+.+\s+(\w+)\s*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := tempRegexp[ostype].FindStringSubmatch(line); matches != nil {
			x := EnvironmentItem{
				Name:        strings.TrimSpace(matches[1] + " " + matches[2]),
				IsTemp:      true,
				Temperature: util.Str2float64(matches[3]),
			}
			items = append(items, x)
		} else if matches := powerRegexp[ostype].FindStringSubmatch(line); matches != nil {
			ok := matches[3] == "Normal" || matches[3] == "good" || matches[3] == "ok"
			x := EnvironmentItem{
				Name:   strings.TrimSpace(matches[1] + " " + matches[2]),
				IsTemp: false,
				OK:     ok,
				Status: matches[3],
			}
			items = append(items, x)
		}
	}
	return items, nil
}
