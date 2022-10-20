package main

import (
	"context"
	"os"

	"github.com/davecgh/go-spew/spew"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
)

var (
	mgr manager.Manager

	log = logf.Log.WithName("controller-examples")
)

func main() {
	logf.SetLogger(zap.New())

	var log = logf.Log.WithName("builder-examples")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr).  // Create the ControllerManagedBy
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

func (a *MachineController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	spew.Dump(req)
	return reconcile.Result{}, nil
}
