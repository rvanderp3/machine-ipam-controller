DBG         ?= 0
#REGISTRY    ?= quay.io/openshift/
VERSION     ?= $(shell git describe --always --abbrev=7)
MUTABLE_TAG ?= latest
IMAGE        = $(REGISTRY)machine-ipam-controller
BUILD_IMAGE ?= registry.ci.openshift.org/openshift/release:golang-1.19
GOLANGCI_LINT = go run ./vendor/github.com/golangci/golangci-lint/cmd/golangci-lint

# Enable go modules and vendoring
# https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support
# https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away
GO111MODULE = on
export GO111MODULE
GOFLAGS ?= -mod=vendor
export GOFLAGS

ifeq ($(DBG),1)
GOGCFLAGS ?= -gcflags=all="-N -l"
endif

.PHONY: all
all: check build test

NO_DOCKER ?= 0

ifeq ($(shell command -v podman > /dev/null 2>&1 ; echo $$? ), 0)
	ENGINE=podman
else ifeq ($(shell command -v docker > /dev/null 2>&1 ; echo $$? ), 0)
	ENGINE=docker
else
	NO_DOCKER=1
endif

USE_DOCKER ?= 0
ifeq ($(USE_DOCKER), 1)
	ENGINE=docker
endif

ifeq ($(NO_DOCKER), 1)
  DOCKER_CMD =
  IMAGE_BUILD_CMD = imagebuilder
else
  DOCKER_CMD := $(ENGINE) run --env GO111MODULE=$(GO111MODULE) --env GOFLAGS=$(GOFLAGS) --rm -v "$(PWD)":/go/src/github.com/rvanderp3/machine-ipam-controller:Z  -w /go/src/github.com/rvanderp3/machine-ipam-controller $(BUILD_IMAGE)
  # The command below is for building/testing with the actual image that Openshift uses. Uncomment/comment out to use instead of above command. CI registry pull secret is required to use this image.
  # DOCKER_CMD := $(ENGINE) run --env GO111MODULE=$(GO111MODULE) --env GOFLAGS=$(GOFLAGS) --rm -v "$(PWD)":/go/src/github.com/rvanderp3/machine-ipam-controller:Z -w /go/src/github.com/rvanderp3/machine-ipam-controller registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.19-openshift-4.11
  IMAGE_BUILD_CMD = $(ENGINE) build
endif

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
ENVTEST = go run ${PROJECT_DIR}/vendor/sigs.k8s.io/controller-runtime/tools/setup-envtest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26

.PHONY: build
build: machine-ipam-controller ## Build binaries

.PHONY: machine-ipam-controller
machine-ipam-controller:
	$(DOCKER_CMD) ./hack/build.sh ## ./hack/go-build.sh machine-api-operator

# Use podman to build the image.
.PHONY: image
image:
	hack/build-image