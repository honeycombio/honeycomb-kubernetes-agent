# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This was borrowed and modified from github.com/thockin/go-build-template

# The binary to build (just the basename).
BIN := honeycomb-kubernetes-agent

# This repo's root import path (under GOPATH).
PKG := github.com/honeycombio/honeycomb-kubernetes-agent

# Where to push the docker image.
REGISTRY ?= honeycombio

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= amd64

# Which OS to build
GOOS ?= linux

# This version-strategy uses git tags to set the version string
VERSION := $(shell cat version.txt | tr -d '\n')
#
# This version-strategy uses a manual value to set the version string
#VERSION := 1.2.3

###
### These variables should not need tweaking.
###

SRC_DIRS := . # directories which hold app source (not vendored)

ALL_ARCH := amd64 arm arm64 ppc64le

IMAGE := $(REGISTRY)/$(BIN)
BASEIMAGE ?= scratch

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

build-%:
	@$(MAKE) --no-print-directory ARCH=$* build

container-%:
	@$(MAKE) --no-print-directory ARCH=$* container

push-%:
	@$(MAKE) --no-print-directory ARCH=$* push

all-build: $(addprefix build-, $(ALL_ARCH))

all-container: $(addprefix container-, $(ALL_ARCH))

all-push: $(addprefix push-, $(ALL_ARCH))

build: bin/$(ARCH)/$(BIN)

bin/$(ARCH)/$(BIN): build-dirs
	@echo "building: $@"
	ARCH=$(ARCH) GOOS=$(GOOS) VERSION=$(VERSION) PKG=$(PKG) BIN=$(BIN) ./build/build.sh

DOTFILE_IMAGE = $(subst :,_,$(subst /,_,$(IMAGE))-$(VERSION))

container: .container-$(DOTFILE_IMAGE) container-name
.container-$(DOTFILE_IMAGE): bin/$(ARCH)/$(BIN) Dockerfile.in
	@sed \
	    -e 's|ARG_BIN|$(BIN)|g' \
	    -e 's|ARG_ARCH|$(ARCH)|g' \
	    -e 's|ARG_FROM|$(BASEIMAGE)|g' \
	    Dockerfile.in > .dockerfile-$(ARCH)
	@docker build -t $(IMAGE):$(VERSION) -f .dockerfile-$(ARCH) .
	@docker images -q $(IMAGE):$(VERSION) > $@

container-name:
	@echo "container: $(IMAGE):$(VERSION)"

push: .push-$(DOTFILE_IMAGE) push-name
.push-$(DOTFILE_IMAGE): .container-$(DOTFILE_IMAGE)
	@docker push $(IMAGE):$(VERSION)
	@docker images -q $(IMAGE):$(VERSION) > $@

push-name:
	@echo "pushed: $(IMAGE):$(VERSION)"

# Push the image tagged with :head
push-head:
	@docker tag $(IMAGE):$(VERSION) $(IMAGE):head
	@docker push $(IMAGE):head
	@echo "pushed: $(IMAGE):head"

version:
	@echo $(VERSION)

test: build-dirs
	./build/test.sh $(SRC_DIRS)                                    \

build-dirs:
	@mkdir -p bin/$(ARCH)

clean: container-clean bin-clean

container-clean:
	rm -rf .container-* .dockerfile-* .push-*

bin-clean:
	rm -rf bin
