#!/bin/bash
set -euxo pipefail

curl -C - -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-linux-amd64 && chmod +x kind
export KUBECONFIG=$(dirname $0)/internal/.kube/config
./kind delete cluster
./kind create cluster --wait 300s

# Build the agent and mock API host images
VERSION=$(cat version.txt | tr -d '\n')

docker build -t apihost:test $(dirname $0)/apihost
docker tag ko.local/honeycomb-kubernetes-agent:$VERSION \
  honeycombio/honeycomb-kubernetes-agent:test
./kind load docker-image apihost:test
./kind load docker-image honeycombio/honeycomb-kubernetes-agent:test

docker build -t internal:test $(dirname $0)/internal
docker run --network=kind --rm --name internal internal:test
