local honeycomb = import "honeycomb-agent-ds.json";
local custom = import "honeycomb-agent-ds-custom.libsonnet";

// Import DaemonSet JSON, append volume to it. The output of this
// equivalent to `honeycomb-agent-ds-custom.json`.
honeycomb +
custom.daemonSet.configVolumeMixin("config")
