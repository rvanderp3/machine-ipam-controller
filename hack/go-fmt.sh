#!/bin/sh
if [ "$IS_CONTAINER" != "" ]; then
  gci_repo=github.com/daixiang0/gci
  for TARGET in "${@}"; do
    find "${TARGET}" -name '*.go' ! -path '*/vendor/*' ! -path '*/.build/*' -exec gofmt -s -w {} \+
    find "${TARGET}" -name '*.go' ! -path '*/vendor/*' ! -path '*/.build/*' -exec go run "$gci_repo" write -s standard -s default -s "prefix(github.com/rvanderp3)" -s blank --skip-generated {} \+
  done
  git diff --exit-code
else
  podman run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/openshift-splat-team/machine-ipam-controller:z" \
    --workdir /go/src/github.com/openshift-splat-team/machine-ipam-controller \
    docker.io/golang:1.19 \
    ./hack/go-fmt.sh "${@}"
fi
