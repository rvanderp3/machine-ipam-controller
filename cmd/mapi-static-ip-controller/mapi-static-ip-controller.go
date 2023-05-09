package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	osclientset "github.com/openshift/client-go/config/clientset/versioned"
	mapiclientset "github.com/openshift/client-go/machine/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ipamcontrollerv1 "github.com/rvanderp3/machine-ipam-controller/pkg/apis/ipamcontroller.openshift.io/v1"
	"github.com/rvanderp3/machine-ipam-controller/pkg/mgmt"
)

var (
	mgr              manager.Manager
	mu               sync.Mutex
	reservedMachines = map[string]struct{}{}
)

func main() {

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Errorf("could not create manager")
		os.Exit(1)
	}
	osclientset.NewForConfig(config.GetConfigOrDie())

	mapiclientset.NewForConfig(config.GetConfigOrDie())

	// Register object scheme to allow deserialization
	machinev1beta1.AddToScheme(mgr.GetScheme())
	ipamcontrollerv1.AddToScheme(mgr.GetScheme())

	err = builder.
		ControllerManagedBy(mgr). // Create the ControllerManagedBy
		For(&machinev1beta1.IPAddressClaim{}).
		Complete(&IPPoolClaimProcessor{})
	if err != nil {
		log.Error(err, "could not create claim processor")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr). // Create the ControllerManagedBy
		For(&ipamcontrollerv1.IPPool{}).
		Complete(&IPPoolController{})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

type IPPoolClaimProcessor struct {
	client.Client
}

type IPPoolController struct {
	client.Client
}

func (a *IPPoolClaimProcessor) BindClaim(ctx context.Context, ipAddressClaim *machinev1beta1.IPAddressClaim) error {
	log.Info("Received BindClaim")
	ip, err := mgmt.GetIPAddress(ctx, ipAddressClaim)
	if err != nil {
		log.Errorf("Unable to get IPAddress: %v", err)
		return err
	}
	log.Infof("Got IPAddress %v", ip)

	// create ipaddress object
	if err = a.Client.Create(ctx, ip); err != nil {
		log.Errorf("Unable to create IPAddress: %v", err)
		return err
	}

	// Update status
	if err = a.Client.Status().Update(ctx, ipAddressClaim); err != nil {
		log.Errorf("Unable to update claim: %v", err)
		return err
	}

	log.Infof("IAC: %v", ipAddressClaim)
	return nil
}

func (a *IPPoolClaimProcessor) ReleaseClaim(ctx context.Context, namespacedName types.NamespacedName) error {
	log.Info("Received ReleaseClaim")
	ipAddress := &machinev1beta1.IPAddress{}
	if err := a.Get(ctx, namespacedName, ipAddress); err != nil {
		return err
	}
	log.Infof("Got IPAddress %v (%v)", ipAddress.Name, ipAddress.Spec.Address)
	if err := mgmt.ReleaseIPConfiguration(ctx, ipAddress); err != nil {
		log.Warnf("Unable to release IP: %v", err)
		return err
	}
	log.Infof("Deleting ipaddress CR %v", ipAddress.Name)
	err := a.Delete(ctx, ipAddress)
	return err
}

func (a *IPPoolClaimProcessor) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	mu.Lock()
	defer mu.Unlock()

	log.Infof("Received request %v", req)

	ipAddressClaim := &machinev1beta1.IPAddressClaim{}
	if err := a.Get(ctx, req.NamespacedName, ipAddressClaim); err != nil {
		log.Warnf("Got error: %v", err)
		if strings.Contains(fmt.Sprintf("%v%", err), "not found") {
			log.Info("Handling remove of claim")
			a.ReleaseClaim(ctx, req.NamespacedName)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	log.Infof("Got IPAddressClaim %v", ipAddressClaim.Name)

	// Check claim to see if it needs IP from a pool that we own.
	poolRef := ipAddressClaim.Spec.PoolRef
	log.Debugf("Kind(%v) Group(%v) Name(%v)", poolRef.Kind, *poolRef.APIGroup, poolRef.Name)
	if poolRef.Kind == ipamcontrollerv1.IPPoolKind && *poolRef.APIGroup == ipamcontrollerv1.APIGroupName {
		log.Debugf("Found a claim for an IP from this provider.  Status: %v", ipAddressClaim.Status)
		if ipAddressClaim.Status.AddressRef.Name == "" {
			err := a.BindClaim(ctx, ipAddressClaim)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			// Status was set.  Verify address still exists?
			log.Info("Ignoring claim due to address already in status")
		}
	}

	return reconcile.Result{}, nil
}

func (a *IPPoolClaimProcessor) InjectClient(c client.Client) error {
	a.Client = c
	log.Info("Set client for claim processor")
	return nil
}

func (a *IPPoolController) LoadPool(ctx context.Context, pool *ipamcontrollerv1.IPPool) error {
	log.Infof("Loading pool: %v", pool.Name)

	// Initialize pool
	err := mgmt.InitializePool(ctx, pool)
	if err == nil {
		// Let's get all IPAddresses and see what has been already claimed to sync
		// the pool
		options := client.ListOptions{
			Namespace: pool.Namespace,
		}
		ipList := machinev1beta1.IPAddressList{}
		err = a.List(ctx, &ipList, &options)
		for _, ip := range ipList.Items {
			if ip.Spec.PoolRef.Name == pool.Name {
				log.Infof("Found IP: %v", ip.Spec.Address)
				err = mgmt.ClaimIPAddress(ctx, pool, ip)
				if err != nil {
					log.Warnf("An error occurred when trying to claim IP %v: %v", ip.Spec.Address, err)
				} else {
					log.Debugf("IP %v is not part of pool %v", ip.Spec.Address, pool.Name)
				}
			}
		}
	}
	return err
}

func (a *IPPoolController) RemovePool(ctx context.Context, pool string) error {
	log.Infof("Removing pool %v", pool)
	ipAddresses := &machinev1beta1.IPAddressList{}
	err := a.Client.List(ctx, ipAddresses)
	if err != nil {
		log.Warnf("Unable to get IPAddresses: %v", err)
		return err
	}

	log.Info("Searching for linked IPAddresses...")
	for _, ip := range ipAddresses.Items {
		log.Debugf("Checking IPAddress: %v", ip.Name)
		if fmt.Sprintf("%v/%v", ip.Namespace, ip.Spec.PoolRef.Name) == pool {
			log.Infof("Deleting ipaddress CR %v", ip.Name)
			mgmt.ReleaseIPConfiguration(ctx, &ip)
			err = a.Delete(ctx, &ip)
			if err != nil {
				log.Warnf("Error occurred while cleaning up IP: %v", err)
			}
		}
	}

	log.Info("Removing pool from mgmt...")
	err = mgmt.RemovePool(ctx, pool)
	if err != nil {
		log.Warnf("Error removing pool from mgmt: %v", err)
	}
	return err
}

func (a *IPPoolController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	mu.Lock()
	defer mu.Unlock()

	log.Infof("Received request %v", req)

	pool := &ipamcontrollerv1.IPPool{}
	if err := a.Get(ctx, req.NamespacedName, pool); err != nil {
		log.Warnf("Got error: %v", err)
		switch t := err.(type) {
		default:
			log.Infof("Type: %v", t)

		}
		if strings.Contains(fmt.Sprintf("%v", err), "not found") {
			log.Info("Handling remove of claim")
			a.RemovePool(ctx, fmt.Sprintf("%v", req))
			return reconcile.Result{}, nil
		} else {
			return reconcile.Result{}, err
		}
	}
	log.Infof("Got Pool %v", pool.Name)
	if err := a.LoadPool(ctx, pool); err != nil {
		log.Errorf("Unable to load pool: %v", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (a *IPPoolController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}
