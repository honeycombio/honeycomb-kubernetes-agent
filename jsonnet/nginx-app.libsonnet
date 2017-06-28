// CHANGE THIS IMPORT TO POINT TO YOUR LOCAL KSONNET
local k = import "/Users/jyao/heptio/hausdorff-ksonnet/ksonnet.beta.2/k.libsonnet";

// Destructure the imports.
local container = k.extensions.v1beta1.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;
local deployment = k.extensions.v1beta1.deployment;
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;
local volume = deployment.mixin.spec.template.spec.volumesType;
local volumeMount = container.volumeMountsType;

// ----------------------------------------------------------------------------
// Mixins.
// ----------------------------------------------------------------------------

// local honeytail = {
//   deployment:: {
//     // configVolumeMixin takes a volume name and produces a mixin
//     // that will append the Honeycomb agent `ConfigMap` to a
//     // `DaemonSet` (as, e.g., the Honeycomb agent is), and then mount
//     // that `ConfigMap` in the subset of containers in the
//     // `DaemonSet` specified by the predicate `containerSelector`.
//     loggingVolumeMixin(volName, mountPath, containerSelector=function(c) true)::
//       local loggingVol = volume.name(volName) + {emptyDir: {}};
//       local loggingMount = volumeMount.new(volName, mountPath);

//       // Add volume to DaemonSet.
//       deployment.mixin.spec.template.spec.volumes([loggingVol]) +

//       // Add volume mount to every container in the DaemonSet.
//       deployment.mapContainers(
//         function (c)
//           if containerSelector(c)
//           then c + container.volumeMounts([loggingMount])
//           else c)
//   }
// };

// ----------------------------------------------------------------------------
// App.
// ----------------------------------------------------------------------------

local targetPort = 80;

// Application deployment.
local podLabels = {app: "nginx"};

local nginxContainer =
  container.new("nginx", "nginx:1.7.9") +
  container.ports(containerPort.containerPort(targetPort));

local nginxDeployment =
  deployment.new("nginx-deployment", 2, nginxContainer, podLabels);

// Application service.
local nginxService =
  local nginxServicePort = servicePort.tcp(targetPort, targetPort);
  service.new("my-nginx", podLabels, nginxServicePort) +
    service.mixin.spec.type("LoadBalancer") +
    service.mixin.metadata.labels(podLabels);

// k.core.v1.list.new([
//   nginxDeployment +
//     honeytail.deployment.loggingVolumeMixin("logging-vol", "/var/logs"),
//   nginxService
// ])

k.core.v1.list.new([
  nginxDeployment,
  nginxService
])
