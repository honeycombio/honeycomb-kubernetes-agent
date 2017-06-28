local honeycomb = import "honeycomb-agent-ds-base.libsonnet";
local custom = import "honeycomb-agent-ds-custom.libsonnet";
local rbac = import "honeycomb-agent-rbac.libsonnet";

local serviceAccountName = "honeycomb-serviceaccount";
local namespace = "kube-system";

// Import Honeycomb agent DaemonSet, append ConfigMap and volumes to it. The output
// of this is equivalent to `honeycomb-agent-ds-app.json`.
local dameonSet = honeycomb.base("honeycomb-agent-v1.1", serviceAccountName, namespace) +
  custom.daemonSet.configVolumeMixin("config") +
  custom.daemonSet.addHostMountedPodLogs("varlog", "varlibdockercontainers");

local rbacComponents = rbac.getRbacComponents(serviceAccountName, namespace);

// NOTE: rbacComponents doesn't work due to https://github.com/ksonnet/ksonnet-lib/issues/43
dameonSet // + rbacComponents
