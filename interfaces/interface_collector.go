package interfaces

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_interface_"

var (
	receiveBytesDesc       *prometheus.Desc
	receiveErrorsDesc      *prometheus.Desc
	receiveDropsDesc       *prometheus.Desc
	receiveBroadcastDesc   *prometheus.Desc
	receiveMulticastDesc   *prometheus.Desc
	transmitBytesDesc      *prometheus.Desc
	transmitErrorsDesc     *prometheus.Desc
	transmitDropsDesc      *prometheus.Desc
	adminStatusDesc        *prometheus.Desc
	operStatusDesc         *prometheus.Desc
	errorStatusDesc        *prometheus.Desc
)

func init() {
	l := []string{"target", "name", "description", "mac", "speed"}
	receiveBytesDesc = prometheus.NewDesc(prefix+"receive_bytes", "Received data in bytes", l, nil)
	receiveErrorsDesc = prometheus.NewDesc(prefix+"receive_errors", "Number of errors caused by incoming packets", l, nil)
	receiveDropsDesc = prometheus.NewDesc(prefix+"receive_drops", "Number of dropped incoming packets", l, nil)
	receiveBroadcastDesc = prometheus.NewDesc(prefix+"receive_broadcast", "Received broadcast packets", l, nil)
	receiveMulticastDesc = prometheus.NewDesc(prefix+"receive_multicast", "Received multicast packets", l, nil)
	transmitBytesDesc = prometheus.NewDesc(prefix+"transmit_bytes", "Transmitted data in bytes", l, nil)
	transmitErrorsDesc = prometheus.NewDesc(prefix+"transmit_errors", "Number of errors caused by outgoing packets", l, nil)
	transmitDropsDesc = prometheus.NewDesc(prefix+"transmit_drops", "Number of dropped outgoing packets", l, nil)
	adminStatusDesc = prometheus.NewDesc(prefix+"admin_up", "Admin operational status", l, nil)
	operStatusDesc = prometheus.NewDesc(prefix+"up", "Interface operational status", l, nil)
	errorStatusDesc = prometheus.NewDesc(prefix+"error_status", "Admin and operational status differ", l, nil)
}

type interfaceCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &interfaceCollector{}
}

// Name returns the name of the collector
func (*interfaceCollector) Name() string {
	return "Interfaces"
}

// Describe describes the metrics
func (*interfaceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- receiveBytesDesc
	ch <- receiveErrorsDesc
	ch <- receiveDropsDesc
	ch <- receiveBroadcastDesc
	ch <- receiveMulticastDesc
	ch <- transmitBytesDesc
	ch <- transmitDropsDesc
	ch <- transmitErrorsDesc
	ch <- adminStatusDesc
	ch <- operStatusDesc
	ch <- errorStatusDesc
}

// Collect collects metrics from Cisco
func (c *interfaceCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show interface")
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse interfaces for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}
	if client.OSType == rpc.IOSXE {
		out, err := client.RunCommand("show vlans")
		if err != nil {
			return err
		}
		vlans, err := c.ParseVlans(client.OSType, out)
		if err != nil {
			if client.Debug {
				log.Printf("Parse vlans for %s: %s\n", labelValues[0], err.Error())
			}
			return nil
		}
		for _, vlan := range vlans {
			for i, item := range items {
				if item.Name == vlan.Name {
					items[i].InputBytes = vlan.InputBytes
					items[i].OutputBytes = vlan.OutputBytes
					break
				}
			}
		}
	}

	for _, item := range items {
		l := append(labelValues, item.Name, item.Description, item.MacAddress, item.Speed)

		errorStatus := 0
		if item.AdminStatus != item.OperStatus {
			errorStatus = 1
		}
		adminStatus := 0
		if item.AdminStatus == "up" {
			adminStatus = 1
		}
		operStatus := 0
		if item.OperStatus == "up" {
			operStatus = 1
		}
		ch <- prometheus.MustNewConstMetric(receiveBytesDesc, prometheus.GaugeValue, item.InputBytes, l...)
		ch <- prometheus.MustNewConstMetric(receiveErrorsDesc, prometheus.GaugeValue, item.InputErrors, l...)
		ch <- prometheus.MustNewConstMetric(receiveDropsDesc, prometheus.GaugeValue, item.InputDrops, l...)
		ch <- prometheus.MustNewConstMetric(transmitBytesDesc, prometheus.GaugeValue, item.OutputBytes, l...)
		ch <- prometheus.MustNewConstMetric(transmitErrorsDesc, prometheus.GaugeValue, item.OutputErrors, l...)
		ch <- prometheus.MustNewConstMetric(transmitDropsDesc, prometheus.GaugeValue, item.OutputDrops, l...)
		ch <- prometheus.MustNewConstMetric(receiveBroadcastDesc, prometheus.GaugeValue, item.InputBroadcast, l...)
		ch <- prometheus.MustNewConstMetric(receiveMulticastDesc, prometheus.GaugeValue, item.InputMulticast, l...)
		ch <- prometheus.MustNewConstMetric(adminStatusDesc, prometheus.GaugeValue, float64(adminStatus), l...)
		ch <- prometheus.MustNewConstMetric(operStatusDesc, prometheus.GaugeValue, float64(operStatus), l...)
		ch <- prometheus.MustNewConstMetric(errorStatusDesc, prometheus.GaugeValue, float64(errorStatus), l...)
	}

	return nil
}
