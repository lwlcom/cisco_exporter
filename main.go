package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/lwlcom/cisco_exporter/config"
	"github.com/lwlcom/cisco_exporter/connector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const version string = "0.2"

var (
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("web.listen-address", ":9362", "Address on which to expose metrics and web interface.")
	metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sshHosts           = flag.String("ssh.targets", "", "SSH Hosts to scrape")
	sshUsername        = flag.String("ssh.user", "cisco_exporter", "Username to use for SSH connection")
	sshPassword        = flag.String("ssh.password", "", "Password to use for SSH connection")
	sshKeyFile         = flag.String("ssh.keyfile", "", "Key file to use for SSH connection")
	sshTimeout         = flag.Int("ssh.timeout", 5, "Timeout to use for SSH connection")
	sshBatchSize       = flag.Int("ssh.batch-size", 10000, "The SSH response batch size")
	debug              = flag.Bool("debug", false, "Show verbose debug output in log")
	legacyCiphers      = flag.Bool("legacy.ciphers", false, "Allow legacy CBC ciphers")
	bgpEnabled         = flag.Bool("bgp.enabled", true, "Scrape bgp metrics")
	environmentEnabled = flag.Bool("environment.enabled", true, "Scrape environment metrics")
	factsEnabled       = flag.Bool("facts.enabled", true, "Scrape system metrics")
	interfacesEnabled  = flag.Bool("interfaces.enabled", true, "Scrape interface metrics")
	opticsEnabled      = flag.Bool("optics.enabled", true, "Scrape optic metrics")
	configFile         = flag.String("config.file", "", "Path to config file")
	devices            []*connector.Device
	cfg                *config.Config
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: cisco_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		log.Fatalf("could not initialize exporter. %v", err)
	}

	startServer()
}

func loadConfig() (*config.Config, error) {
	if len(*configFile) == 0 {
		log.Infoln("Loading config flags")
		return loadConfigFromFlags(), nil
	}

	log.Infoln("Loading config from", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func initialize() error {
	c, err := loadConfig()
	if err != nil {
		return err
	}

	devices, err = devicesForConfig(c)
	if err != nil {
		return err
	}
	cfg = c

	return nil
}

func loadConfigFromFlags() *config.Config {
	c := config.New()

	c.Debug = *debug
	c.LegacyCiphers = *legacyCiphers
	c.Timeout = *sshTimeout
	c.BatchSize = *sshBatchSize
	c.Username = *sshUsername
	c.Password = *sshPassword

	c.KeyFile = *sshKeyFile

	c.DevicesFromTargets(*sshHosts)

	f := c.Features
	f.BGP = bgpEnabled
	f.Environment = environmentEnabled
	f.Facts = factsEnabled
	f.Interfaces = interfacesEnabled
	f.Optics = opticsEnabled

	return c
}

func printVersion() {
	fmt.Println("cisco_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Martin Poppen")
	fmt.Println("Metric exporter for switches and routers running cisco IOS/NX-OS/IOS-XE")
}

func startServer() {
	log.Infof("Starting Cisco exporter (Version: %s)\n", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cisco Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>Cisco Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/lwlcom/cisco_exporter">github.com/lwlcom/cisco_exporter</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()

	c := newCiscoCollector(devices)
	reg.MustRegister(c)

	l := log.New()
	l.Level = log.ErrorLevel

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      l,
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}
