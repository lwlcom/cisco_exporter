package main

import (
	"strings"
	"time"

	"sync"

	"github.com/lwlcom/cisco_exporter/bgp"
	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/lwlcom/cisco_exporter/connector"
	"github.com/lwlcom/cisco_exporter/environment"
	"github.com/lwlcom/cisco_exporter/facts"
	"github.com/lwlcom/cisco_exporter/interfaces"
	"github.com/lwlcom/cisco_exporter/optics"
	"github.com/lwlcom/cisco_exporter/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "cisco_"

var (
	scrapeDurationDesc *prometheus.Desc
	upDesc             *prometheus.Desc
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
	scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one target", []string{"target"}, nil)
}

type ciscoCollector struct {
	targets    []string
	collectors map[string]collector.RPCCollector
}

func newCiscoCollector(targets []string) *ciscoCollector {
	collectors := collectors()
	return &ciscoCollector{targets, collectors}
}

func collectors() map[string]collector.RPCCollector {
	m := map[string]collector.RPCCollector{}

	if *bgpEnabled == true {
		m["bgp"] = bgp.NewCollector()
	}

	if *environmetEnabled == true {
		m["environment"] = environment.NewCollector()
	}

	if *factsEnabled == true {
		m["facts"] = facts.NewCollector()
	}

	if *interfacesEnabled == true {
		m["interfaces"] = interfaces.NewCollector()
	}

	if *opticsEnabled == true {
		m["optics"] = optics.NewCollector()
	}

	return m
}

// Describe implements prometheus.Collector interface
func (c *ciscoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- scrapeDurationDesc

	for _, col := range c.collectors {
		col.Describe(ch)
	}
}

// Collect implements prometheus.Collector interface
func (c *ciscoCollector) Collect(ch chan<- prometheus.Metric) {
	hosts := c.targets
	wg := &sync.WaitGroup{}

	wg.Add(len(hosts))
	for _, h := range hosts {
		go c.collectForHost(strings.Trim(h, " "), ch, wg)
	}

	wg.Wait()
}

func (c *ciscoCollector) collectForHost(host string, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	l := []string{host}

	t := time.Now()
	defer func() {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(t).Seconds(), l...)
	}()

	conn, err := connector.NewSSSHConnection(host, *sshUsername, *sshKeyFile)
	if err != nil {
		log.Errorln(err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}
	defer conn.Close()

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, l...)

	rpc := rpc.NewClient(conn, *debug)
	err = rpc.Identify()
	if err != nil {
		log.Errorln(host + ": " + err.Error())
		return
	}

	for k, col := range c.collectors {
		err = col.Collect(rpc, ch, l)
		if err != nil && err.Error() != "EOF" {
			log.Errorln(k + ": " + err.Error())
		}
	}
}
