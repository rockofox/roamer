package configuration

// ClusterConfiguration allows the user to override global memory/cpu limits
type ClusterConfiguration struct {
	Memory *int `hcl:"memory"`
	CPU    *int `hcl:"cpu"`
	SafetyMargin *int `hcl:"safety_margin"`
}
