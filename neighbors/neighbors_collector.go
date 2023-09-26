package neighbors

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_neighbors_"

var (
	countDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "name", "protocol", "state"}
	countDesc = prometheus.NewDesc(prefix+"count", "Neighbor count (ARP or IPv6 ND) on interface in state", l, nil)
}

type neighborsCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &neighborsCollector{}
}

// Name returns the name of the collector
func (*neighborsCollector) Name() string {
	return "Neighbors"
}

// Describe describes the metrics
func (*neighborsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- countDesc
}

// Collect collects metrics from Cisco
func (c *neighborsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	c.CollectIPv4(client, ch, labelValues)
	c.CollectIPv6(client, ch, labelValues)
	return nil
}

func (c *neighborsCollector) CollectIPv4(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	iflistcmd := "show ip interface brief"
	out, err := client.RunCommand(iflistcmd)

	if err != nil {
		return err
	}
	interfaces, err := c.ParseInterfacesIPv4(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("ParseInterfaces for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	var interfaces_data = make(map[string]*InterfaceNeighors)
	for _, i := range interfaces {
		interfaces_data[i] = &InterfaceNeighors{}
	}

	out, err = client.RunCommand("show arp detail | include via")
	if err != nil {
		return err
	}

	err = c.ParseIPv4Neighbors(client.OSType, out, interfaces_data)
	if err != nil {
		if client.Debug {
			log.Printf("ParseIPv4Neighbors for %s: %s\n", labelValues[0], err)
		}
		return err
	}

	for i, interface_neigbors := range interfaces_data {
		var l []string
		l = append(labelValues, i, "4", "incomplete")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Incomplete), l...)
		l = append(labelValues, i, "4", "reachable")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Reachable), l...)
		l = append(labelValues, i, "4", "stale")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Stale), l...)
		l = append(labelValues, i, "4", "delay")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Delay), l...)
		l = append(labelValues, i, "4", "probe")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Probe), l...)
	}
	return nil
}

func (c *neighborsCollector) CollectIPv6(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	iflistcmd := "show ipv6 interface brief"
	out, err := client.RunCommand(iflistcmd)

	if err != nil {
		return err
	}
	interfaces, err := c.ParseInterfacesIPv6(client.OSType, out)
	if err != nil {
		if client.Debug {
			log.Printf("ParseInterfaces for %s: %s\n", labelValues[0], err.Error())
		}
		return nil
	}

	var interfaces_data = make(map[string]*InterfaceNeighors)
	for _, i := range interfaces {
		interfaces_data[i] = &InterfaceNeighors{}
	}

	out, err = client.RunCommand("show ipv6 neighbors")
	if err != nil {
		return err
	}

	err = c.ParseIPv6Neighbors(client.OSType, out, interfaces_data)
	if err != nil {
		if client.Debug {
			log.Printf("ParseIPv6Neighbors for %s: %s\n", labelValues[0], err)
		}
		return err
	}

	for i, interface_neigbors := range interfaces_data {
		var l []string
		l = append(labelValues, i, "6", "incomplete")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Incomplete), l...)
		l = append(labelValues, i, "6", "reachable")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Reachable), l...)
		l = append(labelValues, i, "6", "stale")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Stale), l...)
		l = append(labelValues, i, "6", "delay")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Delay), l...)
		l = append(labelValues, i, "6", "probe")
		ch <- prometheus.MustNewConstMetric(countDesc, prometheus.GaugeValue, float64(interface_neigbors.Probe), l...)
	}
	return nil
}
