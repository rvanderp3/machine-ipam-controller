package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	IPPoolKind   = "IPPool"
	APIGroupName = "ipamcontroller.openshift.io"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="CIDR",type=string,JSONPath=`.spec.address-cidr`
// +kubebuilder:printcolumn:name="Prefix",type=integer,JSONPath=`.spec.prefix`
// +kubebuilder:printcolumn:name="Gateway",type=string,JSONPath=`.spec.gateway`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// IPPool represents the IPPool definition for static IPs used by the IPAM controller
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec IPPoolSpec `json:"spec"`

	// status represents the current information/status for the IP pool.
	// Populated by the system.
	// Read-only.
	// +optional
	Status IPPoolStatus `json:"status,omitempty"`
}

// IPPoolSpec is the spec for an IPPool
type IPPoolSpec struct {
	// AddressCidr is a cidr for the IP IPv4range to manage.
	AddressCidr string `json:"address-cidr"`

	// Prefix is the subnet prefix
	Prefix int `json:"prefix"`

	// +optional
	Gateway string `json:"gateway"`

	// +optional
	Nameserver []string `json:"nameserver"`
}

// IPPoolStatus is the current status of an IPPool.
type IPPoolStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []IPPool `json:"items"`
}
