package interfaces

type Interface struct {
	Name        string
	MacAddress  string
	Description string

	AdminStatus string
	OperStatus  string

	InputErrors  float64
	OutputErrors float64

	InputDrops  float64
	OutputDrops float64

	InputBytes  float64
	OutputBytes float64

	InputBroadcast float64
	InputMulticast float64

	Speed string
}
