#!/bin/bash
set -euo pipefail

# Add workdir to path, since minikube-bootstrap.sh installs kubectl/minikube
# there in CI
PATH=$PATH:$(pwd)

# Build the agent and mock API host images
VERSION=$(git describe --tags --always --dirty)
make container

docker build -t apihost $(dirname $0)/apihost
docker tag honeycombio/honeycomb-kubernetes-agent:$VERSION honeycombio/honeycomb-kubernetes-agent:test

# Make them available inside minikube
MK_SSH=$(minikube ssh-key)
set +e
docker save honeycombio/honeycomb-kubernetes-agent:test | ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no -o LogLevel=quiet \
    -i ${MK_SSH} docker load


docker save apihost | ssh -o UserKnownHostsFile=/dev/null \
    -o StrictHostKeyChecking=no -o LogLevel=quiet \
    -i ${MK_SSH} docker load
set -e

kubectl config set-context minikube

# Configure the agent, a basic nginx service, and a mock Honeycomb API host for
# the agent to write to
kubectl create secret generic -n kube-system honeycomb-writekey --from-literal=key=testkey
kubectl apply -f $(dirname $0)/testspec.yaml

sleep 2

NGINX_URL=$(minikube service nginx-service --url)
API_URL=$(minikube service apihost-service --url)

# Make a request to NGINX, check that the agent sends an event to the mock API
curl $NGINX_URL

sleep 1

ret=$(curl $API_URL)
echo "Events received by mock API host:"
echo $ret
count=$(echo $ret | jq ".kubernetestest | length")
if [ $count -ne 1 ]; then
    echo "Didn't receive expected number of events!"
    echo "agent logs:"
    kubectl logs -n kube-system -l k8s-app=honeycomb-agent
    exit 1
fi
