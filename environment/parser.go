package environment

import (
	"errors"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// Parse parses cli output using textfsm and tries to find all temperature, power & fan related data
func (c *environmentCollector) Parse(ostype string, output string) ([]EnvironmentItem, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
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
	results_fan, err := util.ParseTextfsm(templ_fan, output)
	if err != nil {
		return items, errors.New("Error parsing via templ_fan: " + err.Error())
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
	for _, result := range results_fan {
		location := result["LOCATION"].(string)
		name := result["NAME"].(string)
		status := strings.ToLower(strings.TrimSpace(result["STATUS"].(string)))
		status1 := strings.ToLower(strings.TrimSpace(result["STATUS1"].(string)))
		if len(status1) > 0 {
			// parsing power supply fans results in single record with status of two fans
			name = "0"
		}
		status_ok := status == "normal" || status == "good" || status == "ok" || status == "green"
		x := EnvironmentItem{
			Name:   strings.TrimSpace(location + " " + name),
			IsFan:  true,
			OK:     status_ok,
			Status: status,
		}
		items = append(items, x)
		if len(status1) > 0 {
			// add info from the second power supply fan
			name = "1"
			status = status1
			status_ok := status == "normal" || status == "good" || status == "ok" || status == "green"
			x := EnvironmentItem{
				Name:   strings.TrimSpace(location + " " + name),
				IsFan:  true,
				OK:     status_ok,
				Status: status,
			}
			items = append(items, x)
		}
	}
	return items, nil
}
