# Design

The Honeycomb agent is an observability tool that continuously observes
the log files of an application (_e.g._, nginx, MySQL, _etc_.), parses
them, and sends them to the [Honeycomb API][1].

This document describes the core abstractions of the Honeycomb agent,
and how to use it in your Kubernetes cluster. Specifically:

1. What happens when the Honeycomb agent is deployed on a Kubernetes
   cluster
1. How to deploy the Honeycomb agent to a Kubernetes cluster
1. How to configure the agent to consume up logs of different types
   running in your cluster

You will need:

* [Jsonnet][5] v0.9.4
* The [hausdorff:honeycomb][4] branch of ksonnet-lib
* A deployed Kubernetes environment, such as a cluster or minikube.

## Deploying the Honeycomb agent to a Kubernetes cluster (and what happens when you do)

After you have set up a Kubernetes environment (_e.g._, minikube, or a
[cluster on AWS][2]), run the following command.

```shell
kubectl apply -f honeycomb-agent-ds.json
```

This deploys Honeycomb agent is deployed to Kubernetes as a
[`DaemonSet`][3], which means that Kubernetes will try to have the
Honeycomb agent run on every node in the cluster.

When a Honeycomb agent starts, in the general case, it needs to know:

1. Which logs to `tail`
1. Which parsers to use on which logs

The way this works in Kubernetes is: the user annotates some set of
pods with a _label_ (_e.g._, `{app: nginx}`, and then specifies which
parsers to use on pods whose labels match a pattern (specified with
[_label selectors_][7]).

This behavior is specified using a YAML config file, an example of
which is below. An important note about this code is that for the MVP,
this code (which you can see [here][6]) comes _hard-coded_ into the
Honeycomb agent container. This will change to be pluggable as this
project matures.

Here is the example YAML config file:

```yaml
apiHost: https://api.honeycomb.io
writekey: "YOUR_WRITE_KEY_HERE"
watchers:
- dataset: kubernetestest
  parser: json
  sampleRate: 22
  labelSelector: "app=nginx"
- dataset: mysql
  labelSelector: "app=mysql"
  parser: mysql
```

The `watchers` field defines label patterns (called "label selectors")
that tell the Honeycomb agent which pods to read `stdout` of, as well
as which parser to use when it reads the data.

For example, the data set `"kubernetestest"` requires the JSON parser,
and we want Honeycomb to parse the `stdout` of any pod with a label
matching `app=nginx`.

Once the agent boots up, it reads this file to configure itself, and
begins parsing the logs and sending data to Honeycomb. As soon as
there are events to consume, they will begin appearing in the
dashboard.

## Customizing the Honeycomb agent for different applications

When the agents are deployed to the cluster, they need to be
configured so that they have the resources they need to actually get
to the logs of different pods (_e.g._, volume mounts, and so on).

This is typically done by writing a YAML file specifying (_e.g._) a
`DaemonSet`, with all the resources declared, and then `kubectl` this
to the cluster.

But, this is a burden on consumers. Because YAML is purely
declarative, it means that the user must either template the YAML
somehow, or else each configuration of the agent requires a completely
different YAML file to be written.

To reduce this burden, the Honeycomb agent uses [ksonnet][9].

Let's look at an example of how the base Honeycomb agent `DaemonSet`
can be customized to read the container logs for different pods. In
[`honeycomb-agent-ds-app.jsonnet`][8], we see code like the following:

```c++
local honeycomb = import "honeycomb-agent-ds-base.libsonnet";
local custom = import "honeycomb-agent-ds-custom.libsonnet";

// Import Honeycomb agent DaemonSet, append volume to it. The output
// of this equivalent to `honeycomb-agent-ds-custom.json`.
honeycomb.base("honeycomb-agent-v1.1", "kube-system") +
custom.daemonSet.addHostMountedPodLogs("varlog", "varlibdockercontainers")
```

One way to read this code is as saying "take `honeycomb.base`, the
default `DaemonSet` for the Honeycomb agent, and use
`addHostMountedPodLogs` to mount a pods's container logs inside all
the containers in that `DaemonSet`.

Breaking it down, this code is doing three important things:

1. Imports `honeycomb.base`, which will define the most basic
   Honeycomb agent daemonset
1. Imports `custom.daemonSet.addHostMountedPodLogs`, which will
   mount the path to the Pod container logs into a (customizable) subset of the containers in a `DaemonSet`
1. Combining these two things with the Jsonnet `+` ("mixin") operator.

As we can see, this approach allows users to decouple the "base" logic
that all configurations have in common, and customize it as necessary.

Now let's look at how `addHostMountedPodLogs` works.

```c++
    // addhostMountedPodLogs takes a two volume names and produces a
    // mixin that will mount the Kubernetes pod logs into a set of
    // containers specified by `containerSelector`.
    addHostMountedPodLogs(
      varlogVolName, podLogVolName, containerSelector=function(c) true
    )::
      // Pod logs are located on the host at
      // `/var/lib/docker/containers`. Define volumes and mounts for
      // these paths, so the Honeytailer can access them.
      local varlogVol = volume.fromHostPath(varlogVolName, "/var/log");
      local varlogMount =
        volumeMount.new(varlogVol.name, varlogVol.hostPath.path);
      local podLogsVol =
        volume.fromHostPath(
          podLogVolName,
          "/var/lib/docker/containers");
      local podLogMount =
        volumeMount.new(podLogsVol.name, podLogsVol.hostPath.path, true);

      // Add volume to DaemonSet, and attach mounts to every
      // container for which `containerSelector` is true.
      ds.mixin.spec.template.spec.volumes([varlogVol, podLogsVol]) +

      // Add volume mount to every container in the DaemonSet.
      ds.mapContainers(
        function (c)
          if containerSelector(c)
          then c + container.volumeMounts([varlogMount, podLogMount])
          else c),
  }
```

This code is a bit more complicated, but it's essentially doing four
things:

1. Creating a volume that exposes `/var/lib` and
   `/var/lib/docker/containers` from the host. These directories
   contain logs for applications and pod containers. Because they are
   exposed from the host, the Honeycomb agent can simply read from
   those files to get logs for a pod.

   ```c++
   local varlogVol = volume.fromHostPath(varlogVolName, "/var/log");
    local podLogsVol =
      volume.fromHostPath(
        podLogVolName,
        "/var/lib/docker/containers");
   ```
1. Creating volume mounts for both of those volumes, which can be
   embedded in the Honeycomb agent container.

   ```c++
   local varlogMount =
     volumeMount.new(varlogVol.name, varlogVol.hostPath.path);
   local podLogMount =
     volumeMount.new(podLogsVol.name, podLogsVol.hostPath.path, true);
   ```
1. Adding the volumes to the Honeycomb agent pod.
   ```c++
      ds.mixin.spec.template.spec.volumes([varlogVol, podLogsVol]) +
   ```
1. Adding these volume mounts to every container in the Honeycomb agent pod.
   ```c++
   ds.mapContainers(
     function (c)
       if containerSelector(c)
       then c + container.volumeMounts([varlogMount, podLogMount])
       else c),
   ```

## Limitations

* The Honeycomb agent can currently only read `stdout` for container
  logs, but eventually we will expand this to support reading from
  (_e.g._) arbitrary files on a persistent volume.


[1]: https://honeycomb.io/
[2]: https://aws.amazon.com/quickstart/architecture/heptio-kubernetes/
[3]: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[4]: https://github.com/hausdorff/ksonnet-lib/tree/honeycomb
[5]: http://jsonnet.org/
[6]: https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/devel/config.yaml
[7]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
[8]: https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/devel/jsonnet/honeycomb-agent-ds-app.jsonnet
[9]: https://github.com/ksonnet/ksonnet-lib
