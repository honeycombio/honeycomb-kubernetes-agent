#!/bin/bash
set -euo pipefail

export KUBECONFIG=/.kube/config

# get server ip. This used to be hardcoded, but my local docker has a different
# value for this than CircleCI, so ... let's make it check.
ip_prefix=$(cat /etc/hosts | grep -o 172.[0-9]*.0.)
ip_suffix=0
port=6443
set +e
until ( nc -z "${ip_prefix}${ip_suffix}" $port ); do
  echo "tried: ${ip_prefix}${ip_suffix}"
  ((ip_suffix++))
  if [[ "$ip_suffix" -ge 255 ]]; then
    echo "Could not find control plane in ${ip_prefix}0-254."
    exit 1
  fi
done
set -e
echo "Control plane ip: ${ip_prefix}${ip_suffix}."

kubectl config set-cluster kind-kind --server=https://${ip_prefix}${ip_suffix}:$port
kubectl config set-context kind-kind
# Configure the agent, a basic nginx service, and a mock Honeycomb API host for
# the agent to write to
kubectl create secret generic -n kube-system honeycomb-writekey --from-literal=key=testkey
kubectl apply -f /testspec.yaml

kubectl wait --for=condition=available --timeout=30s deployment/nginx-deployment
kubectl wait --for=condition=available --timeout=30s deployment/apihost-deployment
kubectl port-forward svc/nginx-service 9111:80 &
kubectl port-forward svc/apihost-service 9112:5000 &

sleep 15

NGINX_URL=localhost:9111
API_URL=localhost:9112

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
    kubectl logs -n kube-system -l app=honeycomb-agent
    exit 1
fi
kubectl delete pod,svc --all
