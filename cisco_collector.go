package main

import (
	"time"

	"sync"

	"github.com/lwlcom/cisco_exporter/connector"
	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix = "cisco_"

var (
	scrapeCollectorDurationDesc *prometheus.Desc
	scrapeDurationDesc          *prometheus.Desc
	upDesc                      *prometheus.Desc
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
	scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one target", []string{"target"}, nil)
	scrapeCollectorDurationDesc = prometheus.NewDesc(prefix+"collect_duration_seconds", "Duration of a scrape by collector and target", []string{"target", "collector"}, nil)
}

type ciscoCollector struct {
	devices    []*connector.Device
	collectors *collectors
}

func newCiscoCollector(devices []*connector.Device) *ciscoCollector {
	return &ciscoCollector{
		devices:    devices,
		collectors: collectorsForDevices(devices, cfg),
	}
}

// Describe implements prometheus.Collector interface
func (c *ciscoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- scrapeDurationDesc
	ch <- scrapeCollectorDurationDesc

	for _, col := range c.collectors.allEnabledCollectors() {
		col.Describe(ch)
	}
}

// Collect implements prometheus.Collector interface
func (c *ciscoCollector) Collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}

	wg.Add(len(c.devices))
	for _, d := range c.devices {
		go c.collectForHost(d, ch, wg)
	}

	wg.Wait()
}

func (c *ciscoCollector) collectForHost(device *connector.Device, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	l := []string{device.Host}

	t := time.Now()
	defer func() {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(t).Seconds(), l...)
	}()

	conn, err := connector.NewSSSHConnection(device, cfg)
	if err != nil {
		log.Errorln(err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}
	defer conn.Close()

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, l...)

	client := rpc.NewClient(conn, cfg.Debug)
	err = client.Identify()
	if err != nil {
		log.Errorln(device.Host + ": " + err.Error())
		return
	}

	for _, col := range c.collectors.collectorsForDevice(device) {
		ct := time.Now()
		err := col.Collect(client, ch, l)

		if err != nil && err.Error() != "EOF" {
			log.Errorln(col.Name() + ": " + err.Error())
		}

		ch <- prometheus.MustNewConstMetric(scrapeCollectorDurationDesc, prometheus.GaugeValue, time.Since(ct).Seconds(), append(l, col.Name())...)
	}
}
