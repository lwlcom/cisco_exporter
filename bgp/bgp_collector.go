package bgp

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_bgp_session_"

var (
	upDesc               *prometheus.Desc
	receivedPrefixesDesc *prometheus.Desc
	inputMessagesDesc    *prometheus.Desc
	outputMessagesDesc   *prometheus.Desc
)

func init() {
	l := []string{"target", "asn", "ip"}
	upDesc = prometheus.NewDesc(prefix+"up", "Session is up (1 = Established)", l, nil)
	receivedPrefixesDesc = prometheus.NewDesc(prefix+"prefixes_received_count", "Number of received prefixes", l, nil)
	inputMessagesDesc = prometheus.NewDesc(prefix+"messages_input_count", "Number of received messages", l, nil)
	outputMessagesDesc = prometheus.NewDesc(prefix+"messages_output_count", "Number of transmitted messages", l, nil)
}

type bgpCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &bgpCollector{}
}

// Name returns the name of the collector
func (*bgpCollector) Name() string {
	return "BGP"
}

// Describe describes the metrics
func (*bgpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- receivedPrefixesDesc
	ch <- inputMessagesDesc
	ch <- outputMessagesDesc
}

// Collect collects metrics from Cisco
func (c *bgpCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show bgp all summary")
	if err != nil {
		return err
	}
	items, err := c.Parse(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("Parse bgp sessions for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	for _, item := range items {
		l := append(labelValues, item.Asn, item.IP)

		up := 0
		if item.Up {
			up = 1
		}

		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, float64(up), l...)
		ch <- prometheus.MustNewConstMetric(receivedPrefixesDesc, prometheus.GaugeValue, float64(item.ReceivedPrefixes), l...)
		ch <- prometheus.MustNewConstMetric(inputMessagesDesc, prometheus.GaugeValue, float64(item.InputMessages), l...)
		ch <- prometheus.MustNewConstMetric(outputMessagesDesc, prometheus.GaugeValue, float64(item.OutputMessages), l...)
	}

	return nil
}
