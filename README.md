(Under development)

Configuration for Kubernetes cluster level logging to [Honeycomb](https://honeycomb.io).

## Setup
```
kubectl create secret generic honeycomb-writekey --from-literal=key=$WRITEKEY
kubectl create configmap --from-file=td-agent.conf
kubectl create daemonset -f ./fluentd-hny-ds.yml
```

Loosely based on the kubernetes Elasticsearch addon:
https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/fluentd-elasticsearch/fluentd-es-image
