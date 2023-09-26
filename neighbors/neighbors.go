package neighbors

type InterfaceNeighors struct {
	Incomplete float64 `default:"0"`
	Reachable  float64 `default:"0"`
	Stale      float64 `default:"0"`
	Delay      float64 `default:"0"`
	Probe      float64 `default:"0"`
}
