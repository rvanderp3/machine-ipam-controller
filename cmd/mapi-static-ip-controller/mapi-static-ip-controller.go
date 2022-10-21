package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/rvanderp3/machine-ipam-controller/pkg/mgmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	mgr manager.Manager
	mu  sync.Mutex
	log = logf.Log.WithName("ipam-controller")
)

func main() {
	ctx := context.TODO()
	logf.SetLogger(zap.New())

	var log = logf.Log.WithName("builder-examples")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}
	osclientset.NewForConfig(config.GetConfigOrDie())
	mgmt.Initialize(ctx)

	mapiclientset.NewForConfig(config.GetConfigOrDie())
	machinev1beta1.AddToScheme(mgr.GetScheme())
	err = builder.
		ControllerManagedBy(mgr).       // Create the ControllerManagedBy
		For(&machinev1beta1.Machine{}). // ReplicaSet is the Application API
		Complete(&MachineController{})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

type MachineController struct {
	client.Client
}

// RawExtensionFromProviderSpec marshals the machine provider spec.
func RawExtensionFromProviderSpec(spec *machinev1beta1.VSphereMachineProviderSpec) (*runtime.RawExtension, error) {
	if spec == nil {
		return &runtime.RawExtension{}, nil
	}

	var rawBytes []byte
	var err error
	if rawBytes, err = json.Marshal(spec); err != nil {
		return nil, fmt.Errorf("error marshalling providerSpec: %v", err)
	}

	return &runtime.RawExtension{
		Raw: rawBytes,
	}, nil
}

// ProviderSpecFromRawExtension unmarshals the JSON-encoded spec
func ProviderSpecFromRawExtension(rawExtension *runtime.RawExtension) (*machinev1beta1.VSphereMachineProviderSpec, error) {
	if rawExtension == nil {
		return &machinev1beta1.VSphereMachineProviderSpec{}, nil
	}

	spec := new(machinev1beta1.VSphereMachineProviderSpec)
	if err := json.Unmarshal(rawExtension.Raw, &spec); err != nil {
		return nil, fmt.Errorf("error unmarshalling providerSpec: %v", err)
	}

	klog.V(5).Infof("Got provider spec from raw extension: %+v", spec)
	return spec, nil
}

func (a *MachineController) HasMyHook(hooks []machinev1beta1.LifecycleHook) bool {
	myLifecycleHook := mgmt.GetLifecycleHook()
	for _, installedHook := range hooks {
		if installedHook == myLifecycleHook {
			return true
		}
	}
	return false
}

func (a *MachineController) FilterMyHook(hooks []machinev1beta1.LifecycleHook) []machinev1beta1.LifecycleHook {
	myLifecycleHook := mgmt.GetLifecycleHook()

	var lifecycleHooks []machinev1beta1.LifecycleHook
	for _, installedHook := range hooks {
		if installedHook == myLifecycleHook {
			continue
		}
		lifecycleHooks = append(lifecycleHooks, installedHook)
	}

	return lifecycleHooks
}

func (a *MachineController) HandlePreCreate(ctx context.Context, machine *machinev1beta1.Machine) error {
	vsphereProviderSpec, err := ProviderSpecFromRawExtension(machine.Spec.ProviderSpec.Value)
	if err != nil {
		log.Error(err, "unable to parse vSphere provider spec")
		return err
	}
	network := vsphereProviderSpec.Network
	devices := network.Devices
	if len(devices) != 1 {
		log.Error(err, "only a single network adapter is supported")
		return err
	}
	networkConfig, err := mgmt.GetIPConfiguration(ctx)

	if err != nil {
		return err
	}

	vsphereProviderSpec.Network.Devices[0].Config = networkConfig
	machine.Labels["test"] = "1"

	rawExtension, err := RawExtensionFromProviderSpec(vsphereProviderSpec)
	if err != nil {
		return err
	}
	log.Info("setting IP address for machine")
	machine.Spec.ProviderSpec.Value = rawExtension
	err = a.Update(ctx, machine)
	if err != nil {
		return err
	}
	return nil
}

func (a *MachineController) HandlePreTerminate(ctx context.Context, machine *machinev1beta1.Machine) error {
	vsphereProviderSpec, err := ProviderSpecFromRawExtension(machine.Spec.ProviderSpec.Value)
	if err != nil {
		log.Error(err, "unable to parse vSphere provider spec")
		return err
	}
	network := vsphereProviderSpec.Network
	devices := network.Devices
	if len(devices) != 1 {
		log.Error(err, "only a single network adapter is supported")
		return err
	}

	device := devices[0]
	networkConfig := device.Config
	if networkConfig == nil {
		return errors.New("network config not found")
	}
	mgmt.ReleaseIPConfiguration(ctx, networkConfig)

	rawExtension, err := RawExtensionFromProviderSpec(vsphereProviderSpec)
	if err != nil {
		return err
	}
	log.Info("released IP address for machine")
	machine.Spec.ProviderSpec.Value = rawExtension
	err = a.Update(ctx, machine)
	if err != nil {
		return err
	}
	return nil
}

func (a *MachineController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	mu.Lock()
	defer mu.Unlock()
	machine := &machinev1beta1.Machine{}
	if err := a.Get(ctx, req.NamespacedName, machine); err != nil {
		log.Error(err, "unable to get machine")
		return reconcile.Result{}, err
	}

	hooks := machine.Spec.LifecycleHooks
	if len(hooks.PreCreate) > 0 {
		phase := machine.Status.Phase
		if phase == nil || *phase != MACHINE_PHASE_PROVISIONING {
			return reconcile.Result{}, nil
		}
		if a.HasMyHook(hooks.PreCreate) == false {
			return reconcile.Result{}, nil
		}
		log.Info(fmt.Sprintf("machine %s has pre create hook", machine.Name))
		err := a.HandlePreCreate(ctx, machine)
		if err != nil {
			return reconcile.Result{}, err
		}
		log.Info("Removing preCreate hook")
		machine.Spec.LifecycleHooks.PreCreate = a.FilterMyHook(hooks.PreCreate)
		err = a.Update(ctx, machine)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if len(hooks.PreTerminate) > 0 {
		phase := machine.Status.Phase
		if phase == nil || *phase != MACHINE_PHASE_DELETING {
			return reconcile.Result{}, nil
		}
		if a.HasMyHook(hooks.PreTerminate) == false {
			return reconcile.Result{}, nil
		}
		err := a.HandlePreTerminate(ctx, machine)
		if err != nil {
			log.Error(err, "unable to release IP with IP management backend. go to allow terminate.")
		}
		log.Info("Removing preTerminate hook")
		machine.Spec.LifecycleHooks.PreTerminate = a.FilterMyHook(hooks.PreTerminate)
		err = a.Update(ctx, machine)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (a *MachineController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}
