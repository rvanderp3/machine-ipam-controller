package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VSphereMachineProviderSpec is the type that will be embedded in a Machine.Spec.ProviderSpec field
// for an VSphere virtual machine. It is used by the vSphere machine actuator to create a single Machine.
// Compatibility level 2: Stable within a major release for a minimum of 9 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=2
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VSphereMachineProviderSpec struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// UserDataSecret contains a local reference to a secret that contains the
	// UserData to apply to the instance
	// +optional
	UserDataSecret *corev1.LocalObjectReference `json:"userDataSecret,omitempty"`
	// CredentialsSecret is a reference to the secret with vSphere credentials.
	// +optional
	CredentialsSecret *corev1.LocalObjectReference `json:"credentialsSecret,omitempty"`
	// Template is the name, inventory path, or instance UUID of the template
	// used to clone new machines.
	Template string `json:"template"`
	// Workspace describes the workspace to use for the machine.
	// +optional
	Workspace *Workspace `json:"workspace,omitempty"`
	// Network is the network configuration for this machine's VM.
	Network NetworkSpec `json:"network"`
	// NumCPUs is the number of virtual processors in a virtual machine.
	// Defaults to the analogue property value in the template from which this
	// machine is cloned.
	// +optional
	NumCPUs int32 `json:"numCPUs,omitempty"`
	// NumCPUs is the number of cores among which to distribute CPUs in this
	// virtual machine.
	// Defaults to the analogue property value in the template from which this
	// machine is cloned.
	// +optional
	NumCoresPerSocket int32 `json:"numCoresPerSocket,omitempty"`
	// MemoryMiB is the size of a virtual machine's memory, in MiB.
	// Defaults to the analogue property value in the template from which this
	// machine is cloned.
	// +optional
	MemoryMiB int64 `json:"memoryMiB,omitempty"`
	// DiskGiB is the size of a virtual machine's disk, in GiB.
	// Defaults to the analogue property value in the template from which this
	// machine is cloned.
	// This parameter will be ignored if 'LinkedClone' CloneMode is set.
	// +optional
	DiskGiB int32 `json:"diskGiB,omitempty"`
	// Snapshot is the name of the snapshot from which the VM was cloned
	// +optional
	Snapshot string `json:"snapshot"`
	// CloneMode specifies the type of clone operation.
	// The LinkedClone mode is only support for templates that have at least
	// one snapshot. If the template has no snapshots, then CloneMode defaults
	// to FullClone.
	// When LinkedClone mode is enabled the DiskGiB field is ignored as it is
	// not possible to expand disks of linked clones.
	// Defaults to FullClone.
	// When using LinkedClone, if no snapshots exist for the source template, falls back to FullClone.
	// +optional
	CloneMode CloneMode `json:"cloneMode,omitempty"`
}

// CloneMode is the type of clone operation used to clone a VM from a template.
type CloneMode string

const (
	// FullClone indicates a VM will have no relationship to the source of the
	// clone operation once the operation is complete. This is the safest clone
	// mode, but it is not the fastest.
	FullClone CloneMode = "fullClone"
	// LinkedClone means resulting VMs will be dependent upon the snapshot of
	// the source VM/template from which the VM was cloned. This is the fastest
	// clone mode, but it also prevents expanding a VMs disk beyond the size of
	// the source VM/template.
	LinkedClone CloneMode = "linkedClone"
)

// NetworkSpec defines the virtual machine's network configuration.
type NetworkSpec struct {
	// Devices defines the virtual machine's network interfaces.
	Devices []NetworkDeviceSpec `json:"devices"`
}

type DHCPState string

const (
	EnabledState  DHCPState = "Enabled"
	DisabledState DHCPState = "Disabled"
)

// AddressesFromPool is an IPAddressPool that will be used to create
// IPAddressClaims for fulfillment by an external controller.
type AddressesFromPool struct {
	Group string `json:"group"`

	// Kind is the type of resource being referenced
	Resource string `json:"resource"`

	// Name is the name of resource being referenced
	Name string `json:"name"`
}

// NetworkDeviceSpec defines the network configuration for a virtual machine's
// network device.
type NetworkDeviceSpec struct {
	// NetworkName is the name of the vSphere network to which the device
	// will be connected.
	// +optional
	NetworkName string `json:"networkName,omitempty"`

	// DeviceName may be used to explicitly assign a name to the network device
	// as it exists in the guest operating system.
	// +optional
	DeviceName string `json:"deviceName,omitempty"`

	// DHCP4 is a flag that indicates whether or not to use DHCP for IPv4
	// on this device.
	// If enabled, then IPAddrs should not contain any IPv4 addresses and
	// AddressesFromPools must be empty.
	// +optional
	// +kubebuilder:default=Enabled
	// +kubebuilder:validation:Enum=Enabled;Disabled
	DHCP4 DHCPState `json:"dhcp4,omitempty"`

	// DHCP6 is a flag that indicates whether or not to use DHCP for IPv6
	// on this device.
	// If enabled, then IPAddrs should not contain any IPv6 addresses and
	// AddressesFromPools must be empty.
	// +optional
	// +kubebuilder:default=Enabled
	// +kubebuilder:validation:Enum=Enabled;Disabled
	DHCP6 DHCPState `json:"dhcp6,omitempty"`

	// Gateway4 is the IPv4 gateway used by this device.
	// Required when DHCP4 is false.
	// +optional
	// +kubebuilder:validation:Format=ipv4
	Gateway4 string `json:"gateway4,omitempty"`

	// Gateway4 is the IPv4 gateway used by this device.
	// Required when DHCP6 is false.
	// +kubebuilder:validation:Format=ipv6
	// +optional
	Gateway6 string `json:"gateway6,omitempty"`

	// IPAddrs is a list of one or more IPv4 and/or IPv6 addresses and CIDR to assign
	// to this device.  For example: 192.168.1.100/24
	// Required when DHCP4 and DHCP6 are both Disabled.
	// +optional
	IPAddrs []string `json:"ipAddrs,omitempty"`

	// MTU is the device’s Maximum Transmission Unit size in bytes.
	// https://www.rfc-editor.org/rfc/rfc791
	// +optional
	// +kubebuilder:validation:Minimum=576
	// +kubebuilder:validation:Maximum=9000
	// +kubebuilder:validation:Default=1500
	MTU int32 `json:"mtu,omitempty"`

	// MACAddr is the MAC address used by this device.
	// It is generally a good idea to omit this field and allow a MAC address
	// to be generated.
	// Please note that this value must use the VMware OUI to work with the
	// in-tree vSphere cloud provider.
	// +kubebuilder:validation:Pattern=`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	// +optional
	MACAddr string `json:"macAddr,omitempty"`

	// Nameservers is a list of IPv4 and/or IPv6 addresses used as DNS
	// nameservers.
	// Please note that Linux allows only three nameservers (https://linux.die.net/man/5/resolv.conf).
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`

	// Routes is a list of optional, static routes applied to the device.
	// +optional
	Routes []NetworkRouteSpec `json:"routes,omitempty"`

	// SearchDomains is a list of search domains used when resolving IP
	// addresses with DNS.
	// +optional
	SearchDomains []string `json:"searchDomains,omitempty"`

	// AddressesFromPools is a list of references to IP pool types and instances
	// which are handled by an external controller.  The external controller
	// will assign IP addresses in accordance with the IP pool instances.
	AddressesFromPools []AddressesFromPool `json:"addressesFromPool,omitempty"`
}

// NetworkRouteSpec defines a static network route.
type NetworkRouteSpec struct {
	// To is an IPv4 or IPv6 address of the target network.
	// +kubebuilder:validation:Format=ipv4
	To string `json:"to"`
	// Via is the IPv4 or IPv6 address of the gateway.
	// +kubebuilder:validation:Format=ipv4
	// +optional
	Via string `json:"via"`
	// Metric is the weight/priority of the route.
	// +optional
	Metric int32 `json:"metric"`
}

// WorkspaceConfig defines a workspace configuration for the vSphere cloud
// provider.
type Workspace struct {
	// Server is the IP address or FQDN of the vSphere endpoint.
	// +optional
	Server string `gcfg:"server,omitempty" json:"server,omitempty"`
	// Datacenter is the datacenter in which VMs are created/located.
	// +optional
	Datacenter string `gcfg:"datacenter,omitempty" json:"datacenter,omitempty"`
	// Folder is the folder in which VMs are created/located.
	// +optional
	Folder string `gcfg:"folder,omitempty" json:"folder,omitempty"`
	// Datastore is the datastore in which VMs are created/located.
	// +optional
	Datastore string `gcfg:"default-datastore,omitempty" json:"datastore,omitempty"`
	// ResourcePool is the resource pool in which VMs are created/located.
	// +optional
	ResourcePool string `gcfg:"resourcepool-path,omitempty" json:"resourcePool,omitempty"`
}

// VSphereMachineProviderStatus is the type that will be embedded in a Machine.Status.ProviderStatus field.
// It contains VSphere-specific status information.
// Compatibility level 2: Stable within a major release for a minimum of 9 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=2
type VSphereMachineProviderStatus struct {
	metav1.TypeMeta `json:",inline"`

	// InstanceID is the ID of the instance in VSphere
	// +optional
	InstanceID *string `json:"instanceId,omitempty"`
	// InstanceState is the provisioning state of the VSphere Instance.
	// +optional
	InstanceState *string `json:"instanceState,omitempty"`
	// Conditions is a set of conditions associated with the Machine to indicate
	// errors or other status
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// TaskRef is a managed object reference to a Task related to the machine.
	// This value is set automatically at runtime and should not be set or
	// modified by users.
	// +optional
	TaskRef string `json:"taskRef,omitempty"`
}
