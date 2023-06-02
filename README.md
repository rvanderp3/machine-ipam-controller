# machine-ipam-controller

## Overview
An example of a controller which decorates machine resources with 
nmstate state data.  

## What does it do?
This controller manages IPPools that are referenced by IPAddressClaims. 
When an IPAddressClaim is created, this controller will validate that the
claim contains a IPPoolRef that matches the kind `IPPool` and apigroup 
`ipamcontroller.openshift.io`.  If the claim is a match, the controller
will then verify that an IPPool with the specified name exists.  If the
pool is found, the controller will request an IP from the IPPool, create
an IPAddress for the assigned IP and update the `IPAddressClaim`'s status
to have a reference to the created `IPAddress`.

## Why does this even exist?
This project intends to provide a prototype of the concepts discussed in
https://github.com/rvanderp3/enhancements/tree/static-ip-addresses-vsphere .  

For details on building the installer and machine API changes, see [DEV.md](./DEV.md).

## How do I configure it?
Create a file called `ipam-config.yaml`.  This file defines the IP addresses for the
pool.

~~~yaml
apiVersion: ipamcontroller.openshift.io/v1
kind: IPPool
metadata:
  name: testpool
spec:
  address-cidr: 192.168.101.248/29
  prefix: 23
  gateway: 192.168.100.1
  nameserver:
    - 8.8.8.8
~~~

Most parameters are self-explanatory.  `address-cidr` defines the grouping of IP 
address for the pool to maintain.   `prefix` defines the prefix for the subnet.

Note: Be careful when configuring gateways in dual stack configurations.  Enabling 
gateways for both IPv4 and IPv6 may have undesired effects depending on which gateway
provides connectivity to external networks.

To define the IPAddressClaim in the Machineset, you can follow the following example:
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
              - addressesFromPool:
                  - group: 'ipamcontroller.openshift.io'
                    name: testpool
                    resource: 'IPPool'
                nameservers:
                  - 8.8.8.8
                nameserver: 192.168.1.215
                networkName: lab
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

As the `machineset` is scaled, `machines` are created with the `addressFromPool` 
which will be a reference to the IPPool to get an IP address from.

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
