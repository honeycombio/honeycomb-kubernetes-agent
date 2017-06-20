local k = import "ksonnet.beta.2/k.libsonnet";

local ds = k.extensions.v1beta1.daemonSet;
local container = k.extensions.v1beta1.daemonSet.mixin.spec.template.spec.containersType;
local volume = ds.mixin.spec.template.spec.volumesType;
local keyToPath = volume.mixin.configMap.itemsType;
local volumeMount = container.volumeMountsType;

// ----------------------------------------------------------------------------
// Mixin.
// ----------------------------------------------------------------------------

{
  daemonSet:: {
    // configVolumeMixin takes a volume name and produces a mixin
    // that will append the Honeycomb agent `ConfigMap` to a
    // `DaemonSet` (as, e.g., the Honeycomb agent is), and then mount
    // that `ConfigMap` in the subset of containers in the
    // `DaemonSet` specified by the predicate `containerSelector`.
    configVolumeMixin(volName, containerSelector=function(c) true)::
      local configVol = volume.fromConfigMap(
        volName,
        "honeycomb-agent-config",
        keyToPath.new("td-agent.conf", "td-agent.conf"));
      local configMount = volumeMount.new(volName, "/etc/td-agent/");

      // Add volume to DaemonSet.
      ds.mixin.spec.template.spec.volumes([configVol]) +

      // Add volume mount to every container in the DaemonSet.
      ds.mapContainers(
        function (c)
          if containerSelector(c)
          then c + container.volumeMounts([configMount])
          else c)
  }
}
