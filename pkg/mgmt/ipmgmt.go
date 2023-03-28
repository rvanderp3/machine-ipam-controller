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
var ipamV4Prefix *goipam.Prefix
var ipamV6Prefix *goipam.Prefix
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
	if len(ipamConfig.IpamConfig.Ipv4RangeCidr) > 0 {
		ipamV4Prefix, err = ipam.NewPrefix(ctx, ipamConfig.IpamConfig.Ipv4RangeCidr)
	}
	if len(ipamConfig.IpamConfig.Ipv6RangeCidr) > 0 {
		ipamV6Prefix, err = ipam.NewPrefix(ctx, ipamConfig.IpamConfig.Ipv6RangeCidr)
	}
	if err != nil {
		return nil
	}
	return nil
}

func GetLifecycleHook() v1beta1.LifecycleHook {
	return ipamConfig.IpamConfig.LifecycleHook
}

func GetIPConfiguration(ctx context.Context) (*v1beta1.NetworkDeviceSpec, error) {
	var ipAddrs []string
	if ipamV4Prefix != nil {
		ipAddr, err := ipam.AcquireIP(ctx, ipamV4Prefix.Cidr)
		if err != nil {
			return nil, err
		}
		ipAddrs = append(ipAddrs, fmt.Sprintf("%v/%v", ipAddr.IP.String(), ipamConfig.IpamConfig.Ipv4Prefix))
	}

	if ipamV6Prefix != nil {
		ipAddr, err := ipam.AcquireIP(ctx, ipamV6Prefix.Cidr)
		if err != nil {
			return nil, err
		}
		ipAddrs = append(ipAddrs, fmt.Sprintf("%v/%v", ipAddr.IP.String(), ipamConfig.IpamConfig.Ipv6Prefix))
	}

	networkConfig := v1beta1.NetworkDeviceSpec{
		DHCP4:         v1beta1.DisabledState,
		DHCP6:         v1beta1.DisabledState,
		Gateway4:      ipamConfig.IpamConfig.GatewayIPv4,
		Gateway6:      ipamConfig.IpamConfig.GatewayIPv6,
		IPAddrs:       ipAddrs,
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
