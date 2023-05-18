package ipamcontroller

import (
	"k8s.io/apimachinery/pkg/runtime"

	ipamcontrollerv1 "github.com/rvanderp3/machine-ipam-controller/pkg/apis/ipamcontroller.openshift.io/v1"
)

const (
	// InstallConfigVersion is the version supported by this package.
	// If you bump this, you must also update the list of convertable values in
	// pkg/types/conversion/installconfig.go
	InstallConfigVersion = "v1"
)

//go:generate go run ../../../vendor/sigs.k8s.io/controller-tools/cmd/controller-gen crd:crdVersions=v1 paths=./v1 output:dir=../../../install/

// GroupName defines the API group for ipamcontroller.
const GroupName = "ipamcontroller.openshift.io"

var (
	SchemeBuilder = runtime.NewSchemeBuilder(ipamcontrollerv1.Install)
	// Install is a function which adds every version of this group to a scheme
	Install = SchemeBuilder.AddToScheme
)
