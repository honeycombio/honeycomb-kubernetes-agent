local k = import "/Users/alex/src/go/src/github.com/ksonnet/ksonnet-lib/ksonnet.beta.2/k.libsonnet";

// Destructuring imports.
local ds = k.extensions.v1beta1.daemonSet;
local container = k.extensions.v1beta1.daemonSet.mixin.spec.template.spec.containersType;
local envVar = container.envType;
local volume = ds.mixin.spec.template.spec.volumesType;
local keyToPath = volume.mixin.configMap.itemsType;
local volumeMount = container.volumeMountsType;

// ----------------------------------------------------------------------------
// Honeycomb agent parts. Containers, volumes, etc.
// ----------------------------------------------------------------------------

local honeycombLabels = {
  "k8s-app": "honeycomb-agent",
  "kubernetes.io/cluster-service": "true",
  version: "v1.1",
};

local dsContainer =
  container.new("honeycomb-agent", "honeycombio/honeycomb-kubernetes-agent:1.1") +
  container.mixin.resources.limits({memory: "200Mi"}) +
  container.mixin.resources.requests({memory: "200Mi", cpu: "100m"}) +
  container.env([
    envVar.fromSecretRef("HONEYCOMB_WRITEKEY", "honeycomb-writekey", "key"),
    envVar.new("HONEYCOMB_DATASET", "kubernetes"),
  ]);

// ----------------------------------------------------------------------------
// App definition. Honeycomb agent DaemonDet
// ----------------------------------------------------------------------------

{
  // base takes a name and a namespace and outputs the default
  // DaemonSet for the Honeycomb agent.
  base(name, namespace)::
    ds.new() +
    // Metadata.
    ds.mixin.metadata.name(name) +
    ds.mixin.metadata.namespace(namespace) +
    ds.mixin.metadata.labels(honeycombLabels) +
    // Template.
    ds.mixin.spec.template.metadata.labels(honeycombLabels) +
    ds.mixin.spec.template.spec.containers(dsContainer) +
    ds.mixin.spec.template.spec.terminationGracePeriodSeconds(30)
}
