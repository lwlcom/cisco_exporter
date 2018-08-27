package bgp

type BgpSession struct {
	IP               string
	Asn              string
	Up               bool
	ReceivedPrefixes float64
	InputMessages    float64
	OutputMessages   float64
}
