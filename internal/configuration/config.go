package configuration

// Config provides the structure of the config hcl file
type Config struct {
	ClusterConfig *ClusterConfiguration `hcl:"cluster,block"`
	Groups        []GroupConfig         `hcl:"group,block"`
}
