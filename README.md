# machine-ipam-controller

## Overview
An example of a controller which decorates machine resources with 
nmstate state data.  

## What does it do?
This controller watches machines and checks for the presence of a 
`preCreate` lifecycle hook.  If present, the controller will request 
an IP address from IPAM, generate nmstate and decorate the machine 
resource, remove the `preCreate` hook, and add a `preTerminate` hook. 
When the machine is deleted, the controller watches for the `preTerminate`
hook and releases the IP address associated with a machine that is to be 
deleted.

## Why does this even exist?
This project intends to provide a prototype of the concepts discussed in
https://github.com/rvanderp3/enhancements/tree/static-ip-addresses-vsphere .  

For details on building the installer and machine API changes, see [DEV.md](./DEV.md).

## How do I configure it?
Create a file called `ipam-config.yaml`.  This file defines the basics 
of forming nmstate for machines.

~~~yaml
ipam-config:
  ipv4-range-cidr: 192.168.101.64/28
  ipv4-prefix: 23
  ipv6-range-cidr: 2001::2/64
  ipv6-prefix: 64
  nameserver:
    - 192.168.1.215
  ipv4-gateway: 192.168.100.1
  ipv6-gateway: 2001::100
  lifecycle-hook:
    name: ipamController
    owner: network-admin
~~~

Most parameters are self-explanatory, `lifecycle-hook` defines the lifecycle
hook associated with this controller.  IPv4 and/or IPv6 addreses may be provisioned
by the controller.  To disable IPv4 or IPv6, comment out the related fields in the 
configuration.  

Note: Be careful when configuring gateways in dual stack configurations.  Enabling 
gateways for both IPv4 and IPv6 may have undesired effects depending on which gateway
provides connectivity to external networks.

`machinesets` which should have static IPs applied should be annotated with 
`preProvision` lifecycle hook matching the hook that is defined here.

For example:
~~~yaml
apiVersion: machine.openshift.io/v1beta1
kind: MachineSet
metadata:
  name: static-machineset-worker
  namespace: openshift-machine-api
  labels:
    machine.openshift.io/cluster-api-cluster: cluster
spec:
  replicas: 0
  selector:
    matchLabels:
      machine.openshift.io/cluster-api-cluster: cluster
      machine.openshift.io/cluster-api-machineset: static-machineset-worker
  template:
    metadata:
      labels:
        machine.openshift.io/cluster-api-cluster: cluster
        machine.openshift.io/cluster-api-machine-role: worker
        machine.openshift.io/cluster-api-machine-type: worker
        machine.openshift.io/cluster-api-machineset: static-machineset-worker
    spec:
      lifecycleHooks:
        preCreate:
          - name: ipamController
            owner: network-admin
      metadata: {}
      providerSpec:
        value:
          numCoresPerSocket: 4
          diskGiB: 120
          snapshot: ''
          userDataSecret:
            name: worker-user-data
          memoryMiB: 16384
          credentialsSecret:
            name: vsphere-cloud-credentials
          network:
            devices:
              - networkName: lab
          metadata:
            creationTimestamp: null
          numCPUs: 4
          kind: VSphereMachineProviderSpec
          workspace:
            datacenter: datacenter
            datastore: datastore
            folder: /datacenter/vm/folder
            resourcePool: /datacenter/host/cluster/Resources
            server: vcenter.test.net
          template: cluster-rhcos
          apiVersion: machine.openshift.io/v1beta1
~~~

As the `machineset` is scaled, `machines` are created with the `preCreate` lifecycle hook.

## How do I build it?

~~~
go mod vendor
go mod tidy
./hack/build.sh
~~~

## How do I run it?

~~~
export KUBECONFIG=path/to/kubeconfig
./bin/mapi-static-ip-controller
~~~
