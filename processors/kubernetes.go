package processors

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
)

type KubernetesMetadataProcessor struct {
	PodGetter     k8sagent.PodGetter
	UID           types.UID
	ContainerName string
	lastPodData   *v1.Pod
}

func (k *KubernetesMetadataProcessor) Init(options map[string]interface{}) error {
	return nil
}

func (k *KubernetesMetadataProcessor) Process(data map[string]interface{}) {
	pod, ok := k.PodGetter.Get(k.UID)
	if ok {
		k.lastPodData = pod
	} else if k.lastPodData != nil {
		pod = k.lastPodData
	}
	// TODO add container metadata too

	if pod != nil {
		metadata := extractMetadataFromPod(pod, k.ContainerName)
		for k, v := range metadata {
			data["kubernetes."+k] = v
		}
	}
}

type PodMetadata struct {
	Labels             map[string]string
	Name               string
	Namespace          string
	ResourceVersion    string
	UID                string
	NodeName           string
	NodeSelector       map[string]string
	ServiceAccountName string
	Subdomain          string
}

type ContainerMetadata struct {
	Args         string
	Command      string
	Env          string
	Image        string
	Name         string
	Ports        string
	Resources    string
	VolumeMounts string
	WorkingDir   string
}

func extractMetadataFromPod(pod *v1.Pod, containerName string) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["pod.labels"] = pod.Labels
	ret["pod.name"] = pod.Name
	ret["pod.namespace"] = pod.Namespace
	ret["pod.resourceVersion"] = pod.ResourceVersion
	ret["pod.UID"] = string(pod.UID)
	ret["pod.nodeName"] = pod.Spec.NodeName
	ret["pod.nodeSelector"] = pod.Spec.NodeSelector
	ret["pod.serviceAccountName"] = pod.Spec.ServiceAccountName
	ret["pod.subdomain"] = pod.Spec.Subdomain

	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			ret["container.args"] = container.Args
			ret["container.command"] = container.Command
			ret["container.env"] = container.Env
			ret["container.image"] = container.Image
			ret["container.ports"] = container.Ports
			ret["container.VolumeMounts"] = container.VolumeMounts
			ret["container.workingDir"] = container.WorkingDir
		}
	}
	return ret
}
