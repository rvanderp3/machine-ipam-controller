package data

import "github.com/openshift/api/machine/v1beta1"

type IpamConfig struct {
	Ipv4RangeCidr string                `yaml:"ipv4-range-cidr"`
	Ipv6RangeCidr string                `yaml:"ipv6-range-cidr"`
	Ipv4Prefix    int32                 `yaml:"ipv4-prefix"`
	Ipv6Prefix    int32                 `yaml:"ipv6-prefix"`
	NameServer    []string              `yaml:"nameserver"`
	GatewayIPv4   string                `yaml:"ipv4-gateway"`
	GatewayIPv6   string                `yaml:"ipv6-gateway"`
	LifecycleHook v1beta1.LifecycleHook `yaml:"lifecycle-hook"`
}

type IpamConfigSpec struct {
	IpamConfig IpamConfig `yaml:"ipam-config"`
}
