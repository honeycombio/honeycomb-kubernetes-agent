#!/bin/bash
set -euo pipefail

curl -C - -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-linux-amd64 && chmod +x kind
export KUBECONFIG=$(dirname $0)/internal/.kube/config
./kind delete cluster
./kind create cluster --wait 300s --image kindest/node:v1.18.0

# Build the agent and mock API host images
VERSION=$(cat version.txt | tr -d '\n')
make container

docker build -t apihost:test $(dirname $0)/apihost
docker tag honeycombio/honeycomb-kubernetes-agent:$VERSION \
  honeycombio/honeycomb-kubernetes-agent:test
./kind load docker-image apihost:test
./kind load docker-image honeycombio/honeycomb-kubernetes-agent:test

docker build -t internal:test $(dirname $0)/internal
docker run --network=bridge --rm --name internal internal:test
