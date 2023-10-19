package environment

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output and tries to find oll temperature and power related data
/*
# almost every Cisco model has different output
# ISOXE C9500 example
#sh environment all
Sensor List:  Environmental Monitoring
 Sensor                  Location        State           Reading
 PSOC-MB_0: VOUT         R0              Normal          12079 mV
 35215MB2_0: VOU         R0              Normal          898 mV
 Temp: Outlet_A          R0              Normal          28 Celsius
 Temp: UADP_0_8          R0              Normal          38 Celsius
 PSOC-DB_1: VOUT         R0              Normal          4999 mV
 3570DB3_0: VOUT         R0              Normal          1048 mV
 Temp: Coretemp          R0              Normal          30 Celsius
 Temp: OutletDB          R0              Normal          25 Celsius

Power                                                    Fan States
Supply  Model No              Type  Capacity  Status     0     1
------  --------------------  ----  --------  ---------  -----------
PS0     C9K-PWR-650WAC-R      AC    650 W     ok         good  N/A
PS1     C9K-PWR-650WAC-R      AC    650 W     ok         good  N/A

Fan                 Fan States
Tray    Status      0     1     2     3
------  ----------  -----------------------
FM0     ok          good  good  good  good
FM1     ok          good  good  good  good

*/
func (c *environmentCollector) Parse(ostype string, output string) ([]EnvironmentItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show environment' is not implemented for " + ostype)
	}
	items := []EnvironmentItem{}

	results_temp, err := util.ParseTextfsm(templ_temp, output)
	if err != nil {
		return items, errors.New("Error parsing via templ_temp: " + err.Error())
	}
	results_power, err := util.ParseTextfsm(templ_power, output)
	if err != nil {
		return items, errors.New("Error parsing via templ_power: " + err.Error())
	}
	for _, result := range results_temp {
		location := result["LOCATION"].(string)
		sensor := result["SENSOR"].(string)
		state := strings.ToLower(strings.TrimSpace(result["STATE"].(string)))
		state_ok := state == "normal" || state == "good" || state == "ok" || state == "green"
		x := EnvironmentItem{
			Name:        strings.TrimSpace(location + " " + sensor),
			IsTemp:      true,
			OK:          state_ok,
			Status:      state,
			Temperature: util.Str2float64(result["VALUE"].(string)),
		}
		items = append(items, x)
	}
	for _, result := range results_power {
		location := result["LOCATION"].(string)
		model := result["MODEL"].(string)
		status := strings.ToLower(strings.TrimSpace(result["STATUS"].(string)))
		status_ok := status == "normal" || status == "good" || status == "ok" || status == "green"
		x := EnvironmentItem{
			Name:   strings.TrimSpace(location + " " + model),
			IsTemp: false,
			OK:     status_ok,
			Status: status,
		}
		items = append(items, x)
	}
	return items, nil

	tempRegexp := make(map[string]*regexp.Regexp)
	powerRegexp := make(map[string]*regexp.Regexp)
	tempRegexp[rpc.IOSXE], _ = regexp.Compile(`^\s*(?:Temp: )?(?P<sensor>(?:\w+\s?)+)\s+(?P<location>\w+)\s+(?P<state>(?:\w+\s?)+)\s+(?P<value>\d+) Celsius`)
	powerRegexp[rpc.IOSXE], _ = regexp.Compile(`(PS\d+)\s+([\w\-]+)\s+\w+\s+\d+\s\w+\s+(\w+)`)
	tempRegexp[rpc.IOS], _ = regexp.Compile(`^(?P<location>\d+)\s+(?P<sensor>air \w+(?: +\w+)?)\s+(?P<value>\d+)C \(.*\)\s+\w+$`)
	powerRegexp[rpc.IOS], _ = regexp.Compile(`^(\w+)\s+.+\s+(AC) \w+\s+(\w+)\s+\w+\s+.+\s+.+$`)
	tempRegexp[rpc.NXOS], _ = regexp.Compile(`^(?P<location>\d+)\s+(?P<sensor>.+)\s+\d\d?\s+\d\d?\s+(?P<value>\d\d?)\s+\w+\s*$`)
	powerRegexp[rpc.NXOS], _ = regexp.Compile(`^(\d+)\s+.+\s+(AC)\s+.+\s+.+\s+(\w+)\s*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := util.FindNamedMatches(tempRegexp[ostype], line); len(matches) > 0 {
			x := EnvironmentItem{
				Name:        strings.TrimSpace(matches["location"]) + " " + strings.TrimSpace(matches["sensor"]),
				IsTemp:      true,
				Temperature: util.Str2float64(matches["value"]),
			}
			if state, ok := matches["state"]; ok {
				state = strings.ToLower(strings.TrimSpace(state))
				x.OK = state == "normal" || state == "good" || state == "ok" || state == "green"
				x.Status = state
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
