package configuration

type Config struct {
	Infrastructure InfrastructureConfig `hcl:"infrastructure,block"`
	Groups         []GroupConfig        `hcl:"group,block"`
}

type InfrastructureConfig struct {
	Memory       int `hcl:"memory"`
	CPU          int `hcl:"cpu"`
	SafetyMargin int `hcl:"safety_margin,optional"`
}
type GroupConfig struct {
	Name   string `hcl:"name,label"`
	Weight int    `hcl:"weight"`
}
