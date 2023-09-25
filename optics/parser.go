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
	virtualNames := [4]string{"Vlan", "Loopback", "Tunnel", "Port-channel"}
	deviceNameRegexp, _ := regexp.Compile(`^(?:Interface\s)?([a-zA-Z0-9\/\.-]+)(?: is disabled)?\s*$`)
	lines := strings.Split(output, "\n")
LINES:
	for _, line := range lines {
		matches := deviceNameRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		// ignore virtual interfaces
		for _, virtualName := range virtualNames {
			if strings.HasPrefix(matches[1], virtualName) {
				continue LINES
			}
		}
		items = append(items, matches[1])
	}
	return items, nil
}

/* ParseTransceiver parses cli output and tries to find tx/rx power for an interface
 * examples for IOS & IOSXE:
 *
 *                                              Optical   Optical
 *              Temperature  Voltage  Current   Tx Power  Rx Power
 * Port         (Celsius)    (Volts)  (mA)      (dBm)     (dBm)
 * ---------    -----------  -------  --------  --------  --------
 * Te1/1               23.9     3.28      17.6      -5.9      -7.2
 * current             23.9     3.28      17.6      -5.9      -7.2
 * domna               30.9     3.32       0.0       N/A       N/A
 *
 *                                  Optical   Optical
 *            Temperature  Voltage  Tx Power  Rx Power
 * Port       (Celsius)    (Volts)  (dBm)     (dBm)
 * ---------  -----------  -------  --------  --------
 * nocurr            23.9     3.28       1.2     -40.0
 * txna              23.9     3.28       N/A     -40.0
 * tempminus        -23.9     3.28       1.2     -40.0
 * tempna             N/A     3.28       1.2     -40.0
 * voltna            23.9      N/A       1.2     -40.0
 */
func (c *opticsCollector) ParseTransceiver(ostype string, output string) (Optics, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return Optics{}, errors.New("Transceiver data is not implemented for " + ostype)
	}
	transceiverRegexp := make(map[string]*regexp.Regexp)
	transceiverRegexp[rpc.IOS] = regexp.MustCompile(`\S+\s+(?P<temp>(?:-?\d+\.\d+|N\/A))\s+(?P<volt>(?:-?\d+\.\d+|N\/A))(?:\s+(?P<curr>(?:-?\d+\.\d+|N\/A)))?\s+(?P<tx>(?:-?\d+\.\d+|N\/A))\s+(?P<rx>(?:-?\d+\.\d+|N\/A))\s*`)
	transceiverRegexp[rpc.NXOS] = regexp.MustCompile(`\s*Tx Power\s*(?P<tx>(?:-)?\d+\.\d+).*\s*Rx Power\s*(?P<rx>(?:-)?\d+\.\d+).*`)
	transceiverRegexp[rpc.IOSXE] = transceiverRegexp[rpc.IOS]

	matches := transceiverRegexp[ostype].FindStringSubmatch(output)
	if matches == nil {
		return Optics{}, errors.New("Transceiver not found")
	}
	var optics Optics
	for i, name := range transceiverRegexp[ostype].SubexpNames() {
		if i != 0 && name != "" {
			if name == "rx" {
				optics.RxPower = util.Str2float64(matches[i])
			}
			if name == "tx" {
				optics.TxPower = util.Str2float64(matches[i])
			}
		}
	}
	return optics, nil
}
