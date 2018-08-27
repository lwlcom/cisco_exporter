package facts

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/lwlcom/cisco_exporter/util"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *factsCollector) ParseVersion(ostype string, output string) (VersionFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return VersionFact{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.IOSXE], _ = regexp.Compile(`^.*, Version (.+) -.*$`)
	versionRegexp[rpc.IOS], _ = regexp.Compile(`^.*, Version (.+),.*$`)
	versionRegexp[rpc.NXOS], _ = regexp.Compile(`^\s+NXOS: version (.*)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := versionRegexp[ostype].FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return VersionFact{Version: ostype + "-" + matches[1]}, nil
	}
	return VersionFact{}, errors.New("Version string not found")
}

// ParseMemory parses cli output and tries to find current memory usage
func (c *factsCollector) ParseMemory(ostype string, output string) ([]MemoryFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show process memory' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*(\S*) Pool Total:\s*(\d+) Used:\s*(\d+) Free:\s*(\d+)\s*$`)

	items := []MemoryFact{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		item := MemoryFact{
			Type:  matches[1],
			Total: util.Str2float64(matches[2]),
			Used:  util.Str2float64(matches[3]),
			Free:  util.Str2float64(matches[4]),
		}
		items = append(items, item)
	}
	return items, nil
}

// ParseCPU parses cli output and tries to find current CPU utilization
func (c *factsCollector) ParseCPU(ostype string, output string) (CPUFact, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return CPUFact{}, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*CPU utilization for five seconds: (\d+)%\/(\d+)%; one minute: (\d+)%; five minutes: (\d+)%.*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return CPUFact{
			FiveSeconds: util.Str2float64(matches[1]),
			Interrupts:  util.Str2float64(matches[2]),
			OneMinute:   util.Str2float64(matches[3]),
			FiveMinutes: util.Str2float64(matches[4]),
		}, nil
	}
	return CPUFact{}, errors.New("Version string not found")
}
