package environment

type EnvironmentItem struct {
	Name        string
	Status      string
	OK          bool
	IsTemp      bool `default:"false"`
	IsFan       bool `default:"false"`
	Temperature float64
}
