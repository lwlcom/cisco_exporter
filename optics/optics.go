package optics

type Optics struct {
	Name  string
	Index string

	Temp    float64
	TempHAT float64
	TempHWT float64
	TempLWT float64
	TempLAT float64

	Voltage    float64
	VoltageHAT float64
	VoltageHWT float64
	VoltageLWT float64
	VoltageLAT float64

	TxPower    float64
	TxPowerHAT float64
	TxPowerHWT float64
	TxPowerLWT float64
	TxPowerLAT float64

	RxPower    float64
	RxPowerHAT float64
	RxPowerHWT float64
	RxPowerLWT float64
	RxPowerLAT float64

	Lanes map[string]*Optics
}
