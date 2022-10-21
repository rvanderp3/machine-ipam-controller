package mgmt

import (
	"context"
	"errors"
	"net/netip"
	"os"

	"github.com/davecgh/go-spew/spew"
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

func GetIPConfiguration(ctx context.Context) (*v1beta1.NetworkConfig, error) {
	ipAddr, err := ipam.AcquireIP(ctx, ipamPrefix.Cidr)
	if err != nil {
		return nil, err
	}
	networkConfig := v1beta1.NetworkConfig{
		Interfaces: []v1beta1.Interface{
			{
				Name: "ens192",
				IPV4: v1beta1.IPV4Addresses{
					Address: []v1beta1.IPV4Address{
						{
							IP:           ipAddr.IP.String(),
							PrefixLength: ipamConfig.IpamConfig.Prefix,
						},
					},
				},
			},
		},
		DnsResolver: v1beta1.DnsResolver{
			Config: v1beta1.DnsResolverConfig{
				Server: ipamConfig.IpamConfig.NameServer,
			},
		},
		Routes: v1beta1.Routes{
			Config: []v1beta1.RouteConfig{
				{
					NextHopAddress: ipamConfig.IpamConfig.DefaultGateway,
				},
			},
		},
	}
	spew.Dump(networkConfig)
	return &networkConfig, nil
}

func ReleaseIPConfiguration(ctx context.Context, networkConfig *v1beta1.NetworkConfig) error {
	if len(networkConfig.Interfaces) == 0 {
		return errors.New("no interfaces associated with network config")
	}
	iface := networkConfig.Interfaces[0]
	addresses := iface.IPV4.Address
	if len(addresses) == 0 {
		return errors.New("no IP addresses associated with the interface")
	}
	address := addresses[0]
	parsedIP, err := netip.ParseAddr(address.IP)
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
