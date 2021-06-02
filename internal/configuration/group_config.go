package configuration

type GroupConfig struct {
	Name   string `hcl:"name,label"`
	Weight int    `hcl:"weight"`
}
