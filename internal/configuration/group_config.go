package configuration

// GroupConfig represents a group block
type GroupConfig struct {
	Name   string `hcl:"name,label"`
	Weight int    `hcl:"weight"`
}
