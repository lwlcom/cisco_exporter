package environment

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
	temperaturesDesc *prometheus.Desc
	powerSupplyDesc  *prometheus.Desc
)

func init() {
	l := []string{"target", "item"}
	temperaturesDesc = prometheus.NewDesc(prefix+"sensor_temp", "Sensor temperatures", l, nil)
	l = append(l, "status")
	powerSupplyDesc = prometheus.NewDesc(prefix+"power_up", "Status of power supplies (1 OK, 0 Something is wrong)", l, nil)
}

type environmentCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &environmentCollector{}
}

// Describe describes the metrics
func (*environmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperaturesDesc
	ch <- powerSupplyDesc
}

// Collect collects metrics from Cisco
func (c *environmentCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show environment")
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
