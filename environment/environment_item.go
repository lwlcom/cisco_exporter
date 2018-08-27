package environment

type EnvironmentItem struct {
	Name        string
	Status      string
	OK          bool
	IsTemp      bool
	Temperature float64
}
