package util

import (
	"errors"
	"math"
	"regexp"
	"strconv"
)

// Str2float64 converts a string to float64
func Str2float64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return -1
	}
	return value
}

// convert string to float64, but return N/A as NaN
func Str2float64Nan(str string) float64 {
	if str == "N/A" {
		return math.NaN()
	}
	return Str2float64(str)
}

var InterfaceTypes = map[string]string{
	"Fa":                      "FastEthernet",
	"Gi":                      "GigabitEthernet",
	"Te":                      "TenGigabitEthernet",
	"Twe":                     "TwentyFiveGigE",
	"Fo":                      "FortyGigabitEthernet",
	"Hu":                      "HundredGigE",
	"Et":                      "Ethernet",
	"Eth":                     "Ethernet",
	"Vl":                      "Vlan",
	"FD":                      "Fddi",
	"Po":                      "Port-channel",
	"PortCh":                  "Port-channel",
	"Tu":                      "Tunnel",
	"Lo":                      "Loopback",
	"Vi":                      "Virtual-Access",
	"Vt":                      "Virtual-Template",
	"EO":                      "EOBC",
	"Di":                      "Dialer",
	"Se":                      "Serial",
	"PO":                      "POS",
	"PosCh":                   "Pos-channel",
	"Mu":                      "Multilink",
	"AT":                      "ATM",
	"Async":                   "Async",
	"Group-Async":             "Group-Async",
	"MFR":                     "MFR",
	"BRI":                     "BRI",
	"BVI":                     "BVI",
	"Null":                    "Null",
	"Embedded-Service-Engine": "Embedded-Service-Engine",
}

func InterfaceShortToLong(shortName string) (string, error) {
	re, _ := regexp.Compile(`^([^\d+]+)(\d?.*)`)
	match := re.FindStringSubmatch(shortName)
	if match == nil {
		return "", errors.New("Cannot extract short type from given name")
	}
	longName, ok := InterfaceTypes[match[1]]
	if ok {
		return longName + match[2], nil
	}
	return "", errors.New("No long form found for name")
}

// https://stackoverflow.com/a/46202939
// return regexp named capture groups as map
func FindNamedMatches(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, value := range match {
		subMatchMap[r.SubexpNames()[i]] = value
	}

	return subMatchMap
}
