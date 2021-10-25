package optics

import (
	"log"
	"regexp"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_optics_"

var (
	opticsTXDesc *prometheus.Desc
	opticsRXDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "interface"}
	opticsTXDesc = prometheus.NewDesc(prefix+"tx", "Transceiver Tx power", l, nil)
	opticsRXDesc = prometheus.NewDesc(prefix+"rx", "Transceiver Rx power", l, nil)
}

type opticsCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &opticsCollector{}
}

// Name returns the name of the collector
func (*opticsCollector) Name() string {
	return "Optics"
}

// Describe describes the metrics
func (*opticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- opticsTXDesc
	ch <- opticsRXDesc
}

// Collect collects metrics from Cisco
func (c *opticsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var iflistcmd string

	switch client.OSType {
	case rpc.IOS, rpc.IOSXE:
		iflistcmd = "show interfaces stats | exclude disabled"
	case rpc.NXOS:
		iflistcmd = "show interface status | exclude disabled | exclude notconn | exclude sfpAbsent | exclude --------------------------------------------------------------------------------"
	}
	out, err := client.RunCommand(iflistcmd)

	if err != nil {
		return err
	}
	interfaces, err := c.ParseInterfaces(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("ParseInterfaces for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	xeDev, _ := regexp.Compile(`\S(\d+)/(\d+)/(\d+)`)

	for _, i := range interfaces {
		switch client.OSType {
		case rpc.IOS:
			out, err = client.RunCommand("show interfaces " + i + " transceiver")
		case rpc.NXOS:
			out, err = client.RunCommand("show interface " + i + " transceiver details")
		case rpc.IOSXE:
			matches := xeDev.FindStringSubmatch(i)
			if matches == nil {
				continue
			}
			out, err = client.RunCommand("show hw-module subslot " + matches[1] + "/" + matches[2] + " transceiver " + matches[3] + " status")
		}
		if err != nil {
			if client.Debug {
				log.Printf("Transceiver command on %s: %s\n", labelValues[0], err.Error())
			}
			continue
		}
		optic, err := c.ParseTransceiver(client.OSType, out)
		if err != nil {
			if client.Debug {
				log.Printf("Transceiver data for %s: %s\n", labelValues[0], err.Error())
			}
			continue
		}
		l := append(labelValues, i)

		ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, float64(optic.TxPower), l...)
		ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, float64(optic.RxPower), l...)
	}

	return nil
}
