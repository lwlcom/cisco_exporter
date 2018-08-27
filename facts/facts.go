package facts

type VersionFact struct {
	Version string
}

type MemoryFact struct {
	Type  string
	Total float64
	Used  float64
	Free  float64
}

type CPUFact struct {
	FiveSeconds float64
	Interrupts  float64
	OneMinute   float64
	FiveMinutes float64
}
