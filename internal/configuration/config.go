package configuration

type Config struct {
	Infrastructure InfrastructureConfig `hcl:"infrastructure,block"`
	Groups         []GroupConfig        `hcl:"group,block"`
}
