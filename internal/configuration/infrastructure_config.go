package configuration

type InfrastructureConfig struct {
	Memory       int `hcl:"memory"`
	CPU          int `hcl:"cpu"`
	SafetyMargin int `hcl:"safety_margin,optional"`
}
