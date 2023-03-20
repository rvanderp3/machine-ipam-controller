Living document which describes how to build the installer and machine API to support static IPs
for a given point in time.  

Checklist:
- [D] SPLAT-828: [Update the installer platform specification](https://github.com/openshift/installer/pull/6982)
- [D] SPLAT-846: [Apply IP configuration to VM extraconfig to bootstrap/control plane nodes](https://github.com/openshift/installer/pull/6512)
- [D] SPLAT-843: [Update OpenShift API to include vSphere CAPV Static Network Definitions](https://github.com/openshift/api/pull/1338)


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

