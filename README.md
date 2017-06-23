# Cluster-level Kubernetes Logging with Honeycomb

[![Build Status](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent.svg?branch=master)](https://travis-ci.org/honeycombio/honeycomb-kubernetes-agent)

[Honeycomb's](https://honeycomb.io) Kubernetes agent aggregates logs across a Kubernetes cluster. Stop managing log storage in all your clusters and start tracking down real problems.

To learn more, check out the [Honeycomb general quickstart](https://honeycomb.io/get-started/), and [Kubernetes-specific docs](https://honeycomb.io/docs/connect/kubernetes/).

## How it Works

`honeycomb-agent` runs as a [DaemonSet](https://kubernetes.io/docs/admin/daemons/) on each node in a cluster. By default, containers' stdout/stderr are written by the Docker daemon to the node filesystem. `honeycomb-agent` reads and parses these logs, augments them with metadata from the Kubernetes API, and ships them to Honeycomb so that you can see what's going on.

<img src="static/honeycomb-agent.png" alt="architecture diagram" width="75%">

## Quickstart

1. Grab your Honeycomb writekey from your [account page](https://ui.honeycomb.io/account), and store it as a Kubernetes secret:
    ```
    kubectl create secret generic honeycomb-writekey --from-literal=key=$WRITEKEY
    ```

2. Create the logging DaemonSet:
    ```
    kubectl create -f ./honeycomb-agent-ds.yml
    ```

## Production-Ready Use

It's best if all of your containers output structured JSON logs. But that's not
always realistic. In particular, you're likely to operate third-party services,
such as proxies or databases, that don't log JSON.

In order to get useful data from these services, you can use Kubernetes [label
selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
to describe how to parse logs for specific services.

For example, to parse logs from pods with the label `app: nginx` as NGINX logs,
you'd specify the following configuration:

```
parsers:
- labelSelector: "app=nginx"
  parser: nginx
```


### Post-Processing Events

You might want to do additional munging of events before sending them to
Honeycomb. For each label selector, you can specify a list of `processors`,
which will be applied in order. For example:

```
parsers:
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

See the [processor reference](TODO) for a full list of options.


## Development Notes

To test with locally-built images, run `eval $(minikube docker-env)`, then build the image with `docker build -t honeycombio/honeycomb-kubernetes-agent .`. See the [minikube docs](https://github.com/kubernetes/minikube#reusing-the-docker-daemon) for more details on building local images.
