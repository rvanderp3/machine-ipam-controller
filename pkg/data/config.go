package data

import "github.com/openshift/api/machine/v1beta1"

type IpamConfig struct {
	IpRangeCidr    string                `yaml:"ip-range-cidr"`
	Prefix         int32                 `yaml:"prefix"`
	NameServer     []string              `yaml:"nameserver"`
	DefaultGateway string                `yaml:"default-gateway"`
	LifecycleHook  v1beta1.LifecycleHook `yaml:"lifecycle-hook"`
}

type IpamConfigSpec struct {
	IpamConfig IpamConfig `yaml:"ipam-config"`
}
