package main

import (
	"context"
	"os"
	"sync"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/rvanderp3/machine-ipam-controller/pkg/mgmt"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	osclientset "github.com/openshift/client-go/config/clientset/versioned"
	mapiclientset "github.com/openshift/client-go/machine/clientset/versioned"
)

const (
	MACHINE_PHASE_PROVISIONING = "Provisioning"
	MACHINE_PHASE_DELETING     = "Deleting"
)

var (
	mgr              manager.Manager
	mu               sync.Mutex
	reservedMachines = map[string]struct{}{}
)

func main() {
	ctx := context.TODO()

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Errorf("could not create manager")
		os.Exit(1)
	}
	osclientset.NewForConfig(config.GetConfigOrDie())
	mgmt.Initialize(ctx)

	mapiclientset.NewForConfig(config.GetConfigOrDie())
	machinev1beta1.AddToScheme(mgr.GetScheme())
	err = builder.
		ControllerManagedBy(mgr). // Create the ControllerManagedBy
		For(&machinev1beta1.IPAddressClaim{}).
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

type IPPoolController struct {
	client.Client
}

func (a *IPPoolController) BindClaim(ctx context.Context, ipAddressClaim *machinev1beta1.IPAddressClaim) error {
	return nil
}

func (a *IPPoolController) ReleaseClaim(ctx context.Context, ipAddressClaim *machinev1beta1.IPAddressClaim) error {
	return nil
}

func (a *IPPoolController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	mu.Lock()
	defer mu.Unlock()
	ipAddressClaim := &machinev1beta1.IPAddressClaim{}
	if err := a.Get(ctx, req.NamespacedName, ipAddressClaim); err != nil {
		return reconcile.Result{}, err
	}

	poolRef := ipAddressClaim.PoolRef
	if poolRef.APIVersion !=

	return reconcile.Result{}, nil
}

func (a *IPPoolController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}
