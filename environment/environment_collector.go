package environment

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
	temperaturesDesc       *prometheus.Desc
	temperaturesStatusDesc *prometheus.Desc
	powerSupplyDesc        *prometheus.Desc
)

func init() {
	l := []string{"target", "item"}
	temperaturesDesc = prometheus.NewDesc(prefix+"sensor_temp", "Sensor temperatures", l, nil)
	l = append(l, "status")
	temperaturesStatusDesc = prometheus.NewDesc(prefix+"sensor_status", "Status of sensor temperatures (1 OK, 0 Something is wrong)", l, nil)
	powerSupplyDesc = prometheus.NewDesc(prefix+"power_up", "Status of power supplies (1 OK, 0 Something is wrong)", l, nil)
}

type environmentCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &environmentCollector{}
}

// Name returns the name of the collector
func (*environmentCollector) Name() string {
	return "Environment"
}

// Describe describes the metrics
func (*environmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperaturesDesc
	ch <- temperaturesStatusDesc
	ch <- powerSupplyDesc
}

// Collect collects metrics from Cisco
func (c *environmentCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var envcmd string

	switch client.OSType {
	case rpc.IOS, rpc.NXOS:
		envcmd = "show environment"
	case rpc.IOSXE:
		envcmd = "show environment all"
	}
	out, err := client.RunCommand(envcmd)
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse environment for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	for _, item := range items {
		l := append(labelValues, item.Name)
		if item.IsTemp {
			ch <- prometheus.MustNewConstMetric(temperaturesDesc, prometheus.GaugeValue, float64(item.Temperature), l...)
			val := 0
			if item.OK {
				val = 1
			}
			l = append(l, item.Status)
			ch <- prometheus.MustNewConstMetric(temperaturesStatusDesc, prometheus.GaugeValue, float64(val), l...)
		} else {
			val := 0
			if item.OK {
				val = 1
			}
			l = append(l, item.Status)
			ch <- prometheus.MustNewConstMetric(powerSupplyDesc, prometheus.GaugeValue, float64(val), l...)
		}
	}

	return nil
}
