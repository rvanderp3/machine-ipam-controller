#!/bin/sh

if [ "$IS_CONTAINER" != "" ]; then
  set -xe
  go generate ./pkg/apis/ipamcontroller.openshift.io/install.go
  set +ex
  # git diff --exit-code
else
  podman run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/openshift/machine-ipam-controller:z" \
    --workdir /go/src/github.com/openshift/machine-ipam-controller \
    docker.io/golang:1.18 \
    ./hack/verify-codegen.sh "${@}"
fi
