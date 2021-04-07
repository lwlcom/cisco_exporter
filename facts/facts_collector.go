package facts

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_facts_"

var (
	versionDesc        *prometheus.Desc
	memoryTotalDesc    *prometheus.Desc
	memoryUsedDesc     *prometheus.Desc
	memoryFreeDesc     *prometheus.Desc
	cpuOneMinuteDesc   *prometheus.Desc
	cpuFiveSecondsDesc *prometheus.Desc
	cpuInterruptsDesc  *prometheus.Desc
	cpuFiveMinutesDesc *prometheus.Desc
)

func init() {
	l := []string{"target"}
	versionDesc = prometheus.NewDesc(prefix+"version", "Running OS version", append(l, "version"), nil)

	memoryTotalDesc = prometheus.NewDesc(prefix+"memory_total", "Total memory", append(l, "type"), nil)
	memoryUsedDesc = prometheus.NewDesc(prefix+"memory_used", "Used memory", append(l, "type"), nil)
	memoryFreeDesc = prometheus.NewDesc(prefix+"memory_free", "Free memory", append(l, "type"), nil)

	cpuOneMinuteDesc = prometheus.NewDesc(prefix+"cpu_one_minute_percent", "CPU utilization for one minute", l, nil)
	cpuFiveSecondsDesc = prometheus.NewDesc(prefix+"cpu_five_seconds_percent", "CPU utilization for five seconds", l, nil)
	cpuInterruptsDesc = prometheus.NewDesc(prefix+"cpu_interrupt_percent", "Interrupt percentage", l, nil)
	cpuFiveMinutesDesc = prometheus.NewDesc(prefix+"cpu_five_minutes_percent", "CPU utilization for five minutes", l, nil)
}

type factsCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &factsCollector{}
}

// Name returns the name of the collector
func (*factsCollector) Name() string {
	return "Facts"
}

// Describe describes the metrics
func (*factsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- versionDesc
	ch <- memoryTotalDesc
	ch <- memoryUsedDesc
	ch <- memoryFreeDesc
}

// CollectVersion collects version informations from Cisco
func (c *factsCollector) CollectVersion(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show version")
	if err != nil {
		return err
	}
	item, err := c.ParseVersion(client.OSType, out)
	if err != nil {
		return err
	}
	l := append(labelValues, item.Version)
	ch <- prometheus.MustNewConstMetric(versionDesc, prometheus.GaugeValue, 1, l...)
	return nil
}

// CollectMemory collects memory informations from Cisco
func (c *factsCollector) CollectMemory(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show process memory")
	if err != nil {
		return err
	}
	items, err := c.ParseMemory(client.OSType, out)
	if err != nil {
		return err
	}
	for _, item := range items {
		l := append(labelValues, item.Type)
		ch <- prometheus.MustNewConstMetric(memoryTotalDesc, prometheus.GaugeValue, item.Total, l...)
		ch <- prometheus.MustNewConstMetric(memoryUsedDesc, prometheus.GaugeValue, item.Used, l...)
		ch <- prometheus.MustNewConstMetric(memoryFreeDesc, prometheus.GaugeValue, item.Free, l...)
	}
	return nil
}

// CollectCPU collects cpu informations from Cisco
func (c *factsCollector) CollectCPU(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand("show process cpu")
	if err != nil {
		return err
	}
	item, err := c.ParseCPU(client.OSType, out)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(cpuOneMinuteDesc, prometheus.GaugeValue, item.OneMinute, labelValues...)
	ch <- prometheus.MustNewConstMetric(cpuFiveSecondsDesc, prometheus.GaugeValue, item.FiveSeconds, labelValues...)
	ch <- prometheus.MustNewConstMetric(cpuInterruptsDesc, prometheus.GaugeValue, item.Interrupts, labelValues...)
	ch <- prometheus.MustNewConstMetric(cpuFiveMinutesDesc, prometheus.GaugeValue, item.FiveMinutes, labelValues...)
	return nil
}

// Collect collects metrics from Cisco
func (c *factsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	err := c.CollectVersion(client, ch, labelValues)
	if client.Debug && err != nil {
		log.Printf("CollectVersion for %s: %s\n", labelValues[0], err.Error())
	}
	err = c.CollectMemory(client, ch, labelValues)
	if client.Debug && err != nil {
		log.Printf("CollectMemory for %s: %s\n", labelValues[0], err.Error())
	}
	err = c.CollectCPU(client, ch, labelValues)
	if client.Debug && err != nil {
		log.Printf("CollectCPU for %s: %s\n", labelValues[0], err.Error())
	}
	return nil
}
