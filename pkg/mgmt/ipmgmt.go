package mgmt

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	goipam "github.com/metal-stack/go-ipam"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1alpha1"

	v1 "github.com/openshift-splat-team/machine-ipam-controller/pkg/apis/ipamcontroller.openshift.io/v1"
)

type PoolInfo struct {
	IPPool *v1.IPPool
	Prefix *goipam.Prefix
}

var ipam = goipam.New()
var ipams = make(map[string]PoolInfo)

func poolKey(pool *v1.IPPool) string {
	return fmt.Sprintf("%v/%v", pool.Namespace, pool.Name)
}

func InitializePool(ctx context.Context, pool *v1.IPPool) error {
	key := poolKey(pool)

	if ipams[key].IPPool == nil {
		if len(pool.Spec.AddressCidr) > 0 {
			ipamPrefix, err := ipam.NewPrefix(ctx, pool.Spec.AddressCidr)
			if err != nil {
				return err
			}
			log.Infof("Created prefix %v", ipamPrefix)
			ipams[key] = PoolInfo{
				IPPool: pool,
				Prefix: ipamPrefix,
			}
		}
	} else {
		// pool already initialized.  Need to validate nothing changed.
		log.Info("Pool already initialized.")
	}

	return nil
}

func RemovePool(ctx context.Context, pool string) error {
	var err error
	// Remove associated IPAddresses
	ippool := ipams[pool]
	if ippool.IPPool != nil {
		log.Info("Removing Prefix...")
		ips := ippool.Prefix
		_, err = ipam.DeletePrefix(ctx, ips.Cidr)
	}

	// Remove Pool
	ipams[pool] = PoolInfo{}
	return err
}

func ClaimIPAddress(ctx context.Context, pool *v1.IPPool, address ipamv1.IPAddress) error {
	poolInfo := ipams[poolKey(pool)]
	if poolInfo.IPPool == nil {
		return errors.New("pool not initialized")
	}

	_, err := ipam.AcquireSpecificIP(ctx, poolInfo.Prefix.Cidr, address.Spec.Address)
	if err != nil {
		return err
	}
	log.Infof("IP %v has been claimed for pool %v", address.Spec.Address, pool.Name)

	return nil
}

func GetIPAddress(ctx context.Context, ipClaim *ipamv1.IPAddressClaim) (*ipamv1.IPAddress, error) {
	var ipAddrs []string

	poolInfo := ipams[fmt.Sprintf("%v/%v", ipClaim.Namespace, ipClaim.Spec.PoolRef.Name)]
	if poolInfo.IPPool == nil {
		return nil, errors.New("pool not initialized")
	}

	ipAddr, err := ipam.AcquireIP(ctx, poolInfo.Prefix.Cidr)
	if err != nil {
		return nil, err
	}
	ipAddrs = append(ipAddrs, fmt.Sprintf("%v", ipAddr.IP.String()))
	apiGroup := "ipamcontroller.openshift.io"
	ipAddress := ipamv1.IPAddress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipClaim.GetName(),
			Namespace: ipClaim.GetNamespace(),
		},
		Spec: ipamv1.IPAddressSpec{
			Address: ipAddrs[0],
			ClaimRef: corev1.LocalObjectReference{
				Name: ipClaim.GetName(),
			},
			Gateway: poolInfo.IPPool.Spec.Gateway,
			PoolRef: corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     "IPPool",
				Name:     ipClaim.Spec.PoolRef.Name,
			},
			Prefix: poolInfo.IPPool.Spec.Prefix,
		},
	}

	return &ipAddress, nil
}

func ReleaseIPConfiguration(ctx context.Context, ipAddr *ipamv1.IPAddress) error {
	address := ipAddr.Spec.Address
	if address == "" {
		return errors.New("no IP addresses associated with the interface")
	}

	log.Infof("Processing ipaddress %v", address)
	parsedIP, err := netip.ParseAddr(address)
	if err != nil {
		return err
	}
	log.Infof("Converted Addr: %v", parsedIP)

	poolInfo := ipams[fmt.Sprintf("%v/%v", ipAddr.Namespace, ipAddr.Spec.PoolRef.Name)]
	ip := &goipam.IP{
		IP:           parsedIP,
		ParentPrefix: poolInfo.Prefix.Cidr,
	}
	log.Info("Releasing IP from pool")
	_, err = ipam.ReleaseIP(ctx, ip)
	return err
}
