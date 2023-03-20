Living document which describes how to build the installer and machine API to support static IPs
for a given point in time.  

Checklist:
- [D] SPLAT-828: [Update the installer platform specification](https://github.com/openshift/installer/pull/6982)
- [D] SPLAT-846: [Apply IP configuration to VM extraconfig to bootstrap/control plane nodes](https://github.com/openshift/installer/pull/6512)
- [R] SPLAT-843: [Update OpenShift API to include vSphere CAPV Static Network Definitions](https://github.com/openshift/api/pull/1338)
- [ ] SPLAT-847: Generate machine manifests for compute nodes
- [ ] SPLAT-848: Generate machine manifests for control plane nodes
- [R] SPLAT-873: start upstream CAPI enhancement for preCreate lifecycle hook
- [ ] SPLAT-845: Apply IP configuration to VM extraconfig to compute nodes
- [R] SPLAT-841: Update OpenShift API to include preCreate hook

# Building the installer

Within the context of static IPs, the installer is responsible for:
- creating bootstrap and control plane VMs with static IPs
- creating compute machinesets with [vSphere CAPV Static Network Definitions](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/main/apis/v1beta1/types.go#L237-L252). The machine API will render the compute VMs with the


1. Clone the installer repo
2. Apply the following patches in order:
- https://patch-diff.githubusercontent.com/raw/openshift/installer/pull/6982.patch
- https://patch-diff.githubusercontent.com/raw/openshift/installer/pull/6512.patch
3. Update go.mod to use API extensions in [api#1338](https://github.com/openshift/api/pull/1338)
~~~go
replace github.com/openshift/api => github.com/rvanderp3/api v0.0.0-20230314214509-08e7188fa099
~~~
4. Build the installer
~~~sh
./hack/build.sh
~~~

# Building the machine API operator


# End to End Testing

Functions implemented in draft or higher maturity PRs:
- [*] Bootstrap and control plane nodes receive static IPs
- [*] Draft of openshift/api changes
- [ ] Control plane machine manifests reflect static IPs
- [ ] Compute machine manifests reflect static IPs
- [ ] Implementation of preProvision lifecycle hook

Prerequisites:
- [*] Build the installer
- [*] Build the machine API operator
- [*] Build a release image with the updated machine API operator
- [*] Create install-config.yaml with node IP addresses


1. Create manifests
2. Create compute machine manifests with a static IP 
3. Create cluster

# Samples

## Sample platform spec with static IP addresses
~~~yaml
platform:
  vsphere:
    datacenter: vanderlab
    apiVIP: 192.168.100.200
    ingressVIP: 192.168.100.201
    network: "lab-pg"
    defaultDatastore: workloadDatastore
    password: "blahblah"
    cluster: cluster1
    username: administrator@vsphere.local
    vCenter: your.vcenter.net
    hosts:
    - role: bootstrap
      networkDevice:
        ipAddrs:
        - 192.168.100.240/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: control-plane
      networkDevice:
        ipAddrs:
        - 192.168.100.241/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: control-plane
      networkDevice:
        ipAddrs:
        - 192.168.100.242/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: control-plane
      networkDevice:
        ipAddrs:
        - 192.168.100.243/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: compute
      networkDevice:
        ipAddrs:
        - 192.168.100.244/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: compute
      networkDevice:
        ipAddrs:
        - 192.168.100.245/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
    - role: compute
      networkDevice:
        ipAddrs:
        - 192.168.100.246/24
        gateway4: 192.168.100.1
        nameservers:
        - 192.168.1.215
~~~



