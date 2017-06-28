local k = import "ksonnet.beta.2/k.libsonnet";

// Destructure the imports.
local container = k.extensions.v1beta1.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;
local deployment = k.extensions.v1beta1.deployment;
local service = k.core.v1.service;
local servicePort = k.core.v1.service.mixin.spec.portsType;
local volume = deployment.mixin.spec.template.spec.volumesType;
local volumeMount = container.volumeMountsType;

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

k.core.v1.list.new([
  nginxDeployment,
  nginxService
])
