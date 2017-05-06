# Cluster-level Kubernetes Logging with Honeycomb

[![Build Status](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent.svg?branch=master)](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent)

[Honeycomb's](https://honeycomb.io) Kubernetes agent aggregates logs across a Kubernetes cluster. Stop managing log storage in all your clusters and start tracking down real problems.

To learn more, check out [our docs](https://honeycomb.io/docs), and our [Kubernetes-specific docs](https://honeycomb.io/docs/connect/kubernetes/).

## How it Works

`honeycomb-agent` runs as a [DaemonSet](https://kubernetes.io/docs/admin/daemons/) on each node in a cluster. By default, containers' stdout/stderr are written by the Docker daemon to the node filesystem. `honeycomb-agent` reads these logs, augments them with metadata from the Kubernetes API, and ships them to Honeycomb so that you can see what's going on.

<img src="static/honeycomb-agent.png" alt="architecture diagram" width="75%">

## Quickstart

1. Grab your Honeycomb writekey from your [account page](https://ui.honeycomb.io/account), and store it as a Kubernetes secret:
    ```
    kubectl create secret generic honeycomb-writekey --from-literal=key=$WRITEKEY
    ```

2. If you want, edit `honeycomb-agent-ds.yml` and set the `HONEYCOMB_DATASET` environment variable to the dataset name you want (by default, data will go to a dataset named "kubernetes").

3. Create the logging DaemonSet:
    ```
    kubectl create -f ./honeycomb-agent-ds.yml
    ```

## Additional configuration
The Honeycomb agent uses [fluentd](http://www.fluentd.org/) to aggregate and ship logs. You might want to modify its configuration to suit your needs: for example, to add custom parsing of specific logs, or to send different classes of logs to different datasets. To do that:

1. Edit the configuration file `td-agent.conf` with your changes.

2. Create a `ConfigMap` with your updated configuration file:
    ```
    kubectl create configmap honeycomb-agent-config --from-file=td-agent.conf
    ```

3. Create an updated logging DaemonSet that reads the ConfigMap:
    ```
    kubectl create -f ./honeycomb-agent-ds-custom.yml
    ```


## Development Notes

For local development with `minikube`, you'll need to change the `mountPath` from `/var/lib/docker/containers/` to `/mnt/sda1/var/lib/docker/containers`.

To test with locally-built images, run `eval $(minikube docker-env)`, then build the image with `docker build -t honeycombio/fluentd-honeycomb ./fluentd-hny-image`. See the [minikube docs](https://github.com/kubernetes/minikube#reusing-the-docker-daemon) for more details on building local images.

This is loosely based on the kubernetes Elasticsearch addon:
https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/fluentd-elasticsearch/fluentd-es-image
