package mgmt

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/netip"
	"os"

	"github.com/go-yaml/yaml"
	goipam "github.com/metal-stack/go-ipam"
	"github.com/openshift/api/machine/v1beta1"
	"github.com/rvanderp3/machine-ipam-controller/pkg/data"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func GetIPAddress(ctx context.Context, ipPool *v1beta1.IPPool, ipClaim *v1beta1.IPAddressClaim) (*v1beta1.IPAddress, error) {
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

	ipAddress := v1beta1.IPAddress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipClaim.GetName(),
			Namespace: ipClaim.GetNamespace(),
		},
		Address: ipAddrs[0],
		ClaimRef: &corev1.ObjectReference{
			Kind: "IPAddressClaim",
			Name: ipClaim.GetName(),
			UID:  ipClaim.GetUID(),
		},
		Gateway: ipamConfig.IpamConfig.GatewayIPv4,
		PoolRef: &corev1.ObjectReference{
			Kind: "IPPool",
			Name: ipPool.Name,
			UID:  ipPool.UID,
		},
		Prefix: int64(ipamConfig.IpamConfig.Ipv4Prefix),
	}
	ipClaim.Status.AddressRef = &corev1.ObjectReference{
		Kind: "IPAddress",
		Name: ipAddress.Name,
	}
	return &ipAddress, nil
}

func ReleaseIPConfiguration(ctx context.Context, networkConfig, ipClaim *v1beta1.IPAddressClaim) error {
	if len(ipClaim.Finalizers) > 0 {
		logrus.Infof("claim %s has pending finalizers, not releasing", ipClaim.Name)
		return nil
	}

	if ipClaim.Status.AddressRef == nil {
		logrus.Infof("claim %s no bound IP address, nothing to release", ipClaim.Name)
		return nil
	}

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
