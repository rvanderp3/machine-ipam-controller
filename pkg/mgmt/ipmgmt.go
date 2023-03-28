package mgmt

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"

	"github.com/go-yaml/yaml"
	goipam "github.com/metal-stack/go-ipam"
	"github.com/openshift/api/machine/v1beta1"
	"github.com/rvanderp3/machine-ipam-controller/pkg/data"
)

const (
	IPConfigurationFile = "ipam-config.yaml"
)

var ipam goipam.Ipamer
var ipamPrefix *goipam.Prefix
var ipamConfig data.IpamConfigSpec

func Initialize(ctx context.Context) error {
	configRaw, err := os.ReadFile(IPConfigurationFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configRaw, &ipamConfig)
	if err != nil {
		return err
	}
	ipam = goipam.New()
	ipamPrefix, err = ipam.NewPrefix(ctx, ipamConfig.IpamConfig.IpRangeCidr)
	if err != nil {
		return nil
	}
	return nil
}

func GetLifecycleHook() v1beta1.LifecycleHook {
	return ipamConfig.IpamConfig.LifecycleHook
}

func GetIPConfiguration(ctx context.Context) (*v1beta1.NetworkDeviceSpec, error) {
	ipAddr, err := ipam.AcquireIP(ctx, ipamPrefix.Cidr)
	if err != nil {
		return nil, err
	}
	networkConfig := v1beta1.NetworkDeviceSpec{
		DHCP4:    v1beta1.DisabledState,
		DHCP6:    v1beta1.DisabledState,
		Gateway4: ipamConfig.IpamConfig.DefaultGateway,
		IPAddrs: []string{
			fmt.Sprintf("%v/%v", ipAddr.IP.String(), ipamConfig.IpamConfig.Prefix),
		},
		Nameservers:   ipamConfig.IpamConfig.NameServer,
		SearchDomains: nil,
	}
	return &networkConfig, nil
}

func ReleaseIPConfiguration(ctx context.Context, networkConfig *v1beta1.NetworkDeviceSpec) error {
	if len(networkConfig.IPAddrs) == 0 {
		return errors.New("no IP addresses associated with network config")
	}

	addresses := networkConfig.IPAddrs
	if len(addresses) == 0 {
		return errors.New("no IP addresses associated with the interface")
	}
	address := addresses[0]
	parsedIP, err := netip.ParseAddr(address)
	if err != nil {
		return err
	}
	ip := &goipam.IP{
		IP:           parsedIP,
		ParentPrefix: "",
	}
	_, err = ipam.ReleaseIP(ctx, ip)
	return err
}
