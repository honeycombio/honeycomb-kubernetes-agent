local honeycomb = import "honeycomb-agent-ds-base.libsonnet";
local custom = import "honeycomb-agent-ds-custom.libsonnet";

// Import Honeycomb agent DaemonSet, append volume to it. The output
// of this equivalent to `honeycomb-agent-ds-custom.json`.
honeycomb.base("honeycomb-agent-v1.1", "kube-system") +
custom.daemonSet.addHostMountedPodLogs("varlog", "varlibdockercontainers")
  // + custom.daemonSet.configVolumeMixin("config")
