package optics

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// ParseInterfaces parses cli output and returns list of interface names
func (c *opticsCollector) ParseInterfaces(ostype string, output string) ([]string, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return nil, errors.New("'show interfaces stats' is not implemented for " + ostype)
	}
	var items []string
	deviceNameRegexp, _ := regexp.Compile(`^([a-zA-Z0-9\/\.-]+)\s*`)
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

// ParseTransceiver parses cli output and tries to find tx/rx power for an interface
func (c *opticsCollector) ParseTransceiver(ostype string, output string) (Optics, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return Optics{}, errors.New("Transceiver data is not implemented for " + ostype)
	}
	transceiverRegexp := make(map[string]*regexp.Regexp)
	transceiverRegexp[rpc.IOS], _ = regexp.Compile(`\S+\s+(?:(?:-)?\d+\.\d+)\s+(?:(?:-)?\d+\.\d+)\s+((?:-)?\d+\.\d+)\s+((?:-)?\d+\.\d+)\s*`)
	transceiverRegexp[rpc.NXOS], _ = regexp.Compile(`\s*Tx Power\s*((?:-)?\d+\.\d+).*\s*Rx Power\s*((?:-)?\d+\.\d+).*`)
	transceiverRegexp[rpc.IOSXE], _ = regexp.Compile(`\s+Transceiver Tx power\s+= ((?:-)?\d+\.\d+).*\s*Transceiver Rx optical power\s+= ((?:-)?\d+\.\d+).*`)

	matches := transceiverRegexp[ostype].FindStringSubmatch(output)
	if matches == nil {
		return Optics{}, errors.New("Transceiver not found")
	}
	return Optics{
		TxPower: util.Str2float64(matches[1]),
		RxPower: util.Str2float64(matches[2]),
	}, nil
}
