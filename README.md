# Cluster-level Kubernetes Logging with Honeycomb

[![Build Status](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent.svg?branch=master)](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent)

[Honeycomb's](https://honeycomb.io) Kubernetes agent aggregates logs across a Kubernetes cluster. Stop managing log storage in all your clusters and start tracking down real problems.

To get started with Honeycomb, check out the [Honeycomb general quickstart](https://honeycomb.io/docs/get-started/).

## How it Works

`honeycomb-agent` runs as a [DaemonSet](https://kubernetes.io/docs/admin/daemons/) on each node in a cluster. It reads container log files from the node's filesystem, augments them with metadata from the Kubernetes API, and ships them to Honeycomb so that you can see what's going on.

<img src="static/honeycomb-agent.png" alt="architecture diagram" width="75%">

## Quickstart

The following steps will deploy the Honeycomb agent to each node in your cluster, and configure it to process logs from all pods.

1. Grab your Honeycomb writekey from your [account page](https://ui.honeycomb.io/account), and create a Kubernetes secret from it:
    ```
    kubectl create secret generic honeycomb-writekey --from-literal=key=$WRITEKEY --namespace=kube-system
    ```

2.  Run the agent
    ```
    kubectl apply -f examples/quickstart.yaml
    ```
    This will do three things:
    - create a service account for the agent so that it can list pods from the API
    - create a minimal `ConfigMap` containing configuration for the agent
    - create a DaemonSet from the agent.

## Production-Ready Use

### Service-specific parsing

It's best if all of your containers output structured JSON logs. But that's not
always realistic. In particular, you're likely to operate third-party services,
such as proxies or databases, that don't log JSON.

You may also want to aggregate logs from specific services, rather than from
everything that might be running in a cluster.

In order to get usefully structured data from services, you can use Kubernetes [label
selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
to describe how to parse logs for specific services.

For example, to parse logs from pods with the label `app: nginx` as NGINX logs,
you'd specify the following configuration:

```
watchers:
- labelSelector: "app=nginx"
  parser: nginx
```
### Post-Processing Events

You might want to do additional munging of events before sending them to
Honeycomb. For each label selector, you can specify a list of `processors`,
which will be applied in order. For example:

```
watchers:
- labelSelector: "app=nginx"
  parser: nginx
  processors:
  - request_shape:            # Unpack the field "request": "GET /path HTTP/1.x
      field: request            # into its constituent components

  - drop_field:               # Remove the "user_email" field from all events
      field: user_email

  - sample:                   # Sample events: only send one in 20
      type: static
      rate: 20
```

See the [docs](/docs/example-configurations.md) for more examples.
