package optics

import (
	"log"

	"github.com/lwlcom/cisco_exporter/rpc"

	"github.com/lwlcom/cisco_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_optics_"

var (
	opticsTempDesc    *prometheus.Desc
	opticsTempHATDesc *prometheus.Desc
	opticsTempHWTDesc *prometheus.Desc
	opticsTempLATDesc *prometheus.Desc
	opticsTempLWTDesc *prometheus.Desc

	opticsVoltageDesc    *prometheus.Desc
	opticsVoltageHATDesc *prometheus.Desc
	opticsVoltageHWTDesc *prometheus.Desc
	opticsVoltageLATDesc *prometheus.Desc
	opticsVoltageLWTDesc *prometheus.Desc

	opticsTXDesc    *prometheus.Desc
	opticsTXHATDesc *prometheus.Desc
	opticsTXHWTDesc *prometheus.Desc
	opticsTXLATDesc *prometheus.Desc
	opticsTXLWTDesc *prometheus.Desc

	opticsRXDesc    *prometheus.Desc
	opticsRXHATDesc *prometheus.Desc
	opticsRXHWTDesc *prometheus.Desc
	opticsRXLATDesc *prometheus.Desc
	opticsRXLWTDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "interface"}
	opticsTempDesc = prometheus.NewDesc(prefix+"temp", "Transceiver temperature in degrees Celsius", l, nil)
	opticsTempHATDesc = prometheus.NewDesc(prefix+"temp_high_alarm_threshold", "Transceiver temperature high alarm threshold", l, nil)
	opticsTempHWTDesc = prometheus.NewDesc(prefix+"temp_high_warn_threshold", "Transceiver temperature high warning threshold", l, nil)
	opticsTempLATDesc = prometheus.NewDesc(prefix+"temp_low_alarm_threshold", "Transceiver temperature low alarm threshold", l, nil)
	opticsTempLWTDesc = prometheus.NewDesc(prefix+"temp_low_warn_threshold", "Transceiver temperature low warning threshold", l, nil)

	opticsVoltageDesc = prometheus.NewDesc(prefix+"module_voltage", "Transceiver voltage", l, nil)
	opticsVoltageHATDesc = prometheus.NewDesc(prefix+"module_voltage_high_alarm_threshold", "Transceiver voltage high alarm threshold", l, nil)
	opticsVoltageHWTDesc = prometheus.NewDesc(prefix+"module_voltage_high_warn_threshold", "Transceiver voltage high warning threshold", l, nil)
	opticsVoltageLATDesc = prometheus.NewDesc(prefix+"module_voltage_low_alarm_threshold", "Transceiver voltage low alarm threshold", l, nil)
	opticsVoltageLWTDesc = prometheus.NewDesc(prefix+"module_voltage_low_warn_threshold", "Transceiver voltage low warning threshold", l, nil)

	l = append(l, "lane")
	opticsTXDesc = prometheus.NewDesc(prefix+"tx", "Transceiver Tx power", l, nil)
	opticsTXHATDesc = prometheus.NewDesc(prefix+"tx_high_alarm_threshold", "Transceiver tx power high alarm threshold", l, nil)
	opticsTXHWTDesc = prometheus.NewDesc(prefix+"tx_high_warn_threshold", "Transceiver tx power high warning threshold", l, nil)
	opticsTXLATDesc = prometheus.NewDesc(prefix+"tx_low_alarm_threshold", "Transceiver tx power low alarm threshold", l, nil)
	opticsTXLWTDesc = prometheus.NewDesc(prefix+"tx_low_warn_threshold", "Transceiver tx power low warning threshold", l, nil)

	opticsRXDesc = prometheus.NewDesc(prefix+"rx", "Transceiver Rx power", l, nil)
	opticsRXHATDesc = prometheus.NewDesc(prefix+"rx_high_alarm_threshold", "Transceiver rx power high alarm threshold", l, nil)
	opticsRXHWTDesc = prometheus.NewDesc(prefix+"rx_high_warn_threshold", "Transceiver rx power high warning threshold", l, nil)
	opticsRXLATDesc = prometheus.NewDesc(prefix+"rx_low_alarm_threshold", "Transceiver rx power low alarm threshold", l, nil)
	opticsRXLWTDesc = prometheus.NewDesc(prefix+"rx_low_warn_threshold", "Transceiver rx power low warning threshold", l, nil)
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
	ch <- opticsTempDesc
	ch <- opticsTempHATDesc
	ch <- opticsTempHWTDesc
	ch <- opticsTempLATDesc
	ch <- opticsTempLWTDesc

	ch <- opticsVoltageDesc
	ch <- opticsVoltageHATDesc
	ch <- opticsVoltageHWTDesc
	ch <- opticsVoltageLATDesc
	ch <- opticsVoltageLWTDesc

	ch <- opticsTXDesc
	ch <- opticsTXHATDesc
	ch <- opticsTXHWTDesc
	ch <- opticsTXLATDesc
	ch <- opticsTXLWTDesc

	ch <- opticsRXDesc
	ch <- opticsRXHATDesc
	ch <- opticsRXHWTDesc
	ch <- opticsRXLATDesc
	ch <- opticsRXLWTDesc
}

// Collect collects metrics from Cisco
func (c *opticsCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {

	switch client.OSType {
	case rpc.IOS, rpc.IOSXE:
		out, err := client.RunCommand("show interface transceiver detail")
		if err != nil {
			if client.Debug {
				log.Printf("Transceiver command on %s: %s\n", labelValues[0], err.Error())
			}
			return nil
		}
		optics_data, err := c.ParseTransceiverAll(client.OSType, out)
		if err != nil {
			if client.Debug {
				log.Printf("ParseTransceiverAll %s: %s\n", labelValues[0], err.Error())
			}
			return nil
		}

		for i, optics := range optics_data {
			l := append(labelValues, i)
			ch <- prometheus.MustNewConstMetric(opticsTempDesc, prometheus.GaugeValue, float64(optics.Temp), l...)
			ch <- prometheus.MustNewConstMetric(opticsTempHATDesc, prometheus.GaugeValue, float64(optics.TempHAT), l...)
			ch <- prometheus.MustNewConstMetric(opticsTempHWTDesc, prometheus.GaugeValue, float64(optics.TempHWT), l...)
			ch <- prometheus.MustNewConstMetric(opticsTempLATDesc, prometheus.GaugeValue, float64(optics.TempLAT), l...)
			ch <- prometheus.MustNewConstMetric(opticsTempLWTDesc, prometheus.GaugeValue, float64(optics.TempLWT), l...)

			ch <- prometheus.MustNewConstMetric(opticsVoltageDesc, prometheus.GaugeValue, float64(optics.Voltage), l...)
			ch <- prometheus.MustNewConstMetric(opticsVoltageHATDesc, prometheus.GaugeValue, float64(optics.VoltageHAT), l...)
			ch <- prometheus.MustNewConstMetric(opticsVoltageHWTDesc, prometheus.GaugeValue, float64(optics.VoltageHWT), l...)
			ch <- prometheus.MustNewConstMetric(opticsVoltageLATDesc, prometheus.GaugeValue, float64(optics.VoltageLAT), l...)
			ch <- prometheus.MustNewConstMetric(opticsVoltageLWTDesc, prometheus.GaugeValue, float64(optics.VoltageLWT), l...)

			var data map[string]*Optics
			if len(optics.Lanes) > 0 {
				data = optics.Lanes
			} else {
				data = map[string]*Optics{"": optics}
			}

			for _, e := range data {
				l2 := append(l, e.Index)
				ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, float64(e.TxPower), l2...)
				ch <- prometheus.MustNewConstMetric(opticsTXHATDesc, prometheus.GaugeValue, float64(e.TxPowerHAT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsTXHWTDesc, prometheus.GaugeValue, float64(e.TxPowerHWT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsTXLATDesc, prometheus.GaugeValue, float64(e.TxPowerLAT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsTXLWTDesc, prometheus.GaugeValue, float64(e.TxPowerLWT), l2...)

				ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, float64(e.RxPower), l2...)
				ch <- prometheus.MustNewConstMetric(opticsRXHATDesc, prometheus.GaugeValue, float64(e.RxPowerHAT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsRXHWTDesc, prometheus.GaugeValue, float64(e.RxPowerHWT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsRXLATDesc, prometheus.GaugeValue, float64(e.RxPowerLAT), l2...)
				ch <- prometheus.MustNewConstMetric(opticsRXLWTDesc, prometheus.GaugeValue, float64(e.RxPowerLWT), l2...)
			}
		}

	case rpc.NXOS:
		iflistcmd := "show interface status | exclude disabled | exclude notconn | exclude sfpAbsent | exclude --------------------------------------------------------------------------------"
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

		for _, i := range interfaces {
			switch client.OSType {
			case rpc.IOS, rpc.IOSXE:
				out, err = client.RunCommand("show interfaces " + i + " transceiver")
			case rpc.NXOS:
				out, err = client.RunCommand("show interface " + i + " transceiver details")
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
					log.Printf("Transceiver data for %s %s: %s\n", labelValues[0], i, err.Error())
				}
				continue
			}
			l := append(labelValues, i)

			ch <- prometheus.MustNewConstMetric(opticsTXDesc, prometheus.GaugeValue, float64(optic.TxPower), l...)
			ch <- prometheus.MustNewConstMetric(opticsRXDesc, prometheus.GaugeValue, float64(optic.RxPower), l...)
		}
	}

	return nil
}
