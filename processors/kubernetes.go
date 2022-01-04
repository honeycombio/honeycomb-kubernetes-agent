package processors

import (
	"regexp"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type KubernetesMetadataProcessor struct {
	PodGetter   k8sagent.PodWatcher
	UID         types.UID
	lastPodData *v1.Pod
}

func (k *KubernetesMetadataProcessor) Init(options map[string]interface{}) error {
	return nil
}

func (k *KubernetesMetadataProcessor) Process(ev *event.Event) bool {
	pod, ok := k.PodGetter.Get(k.UID)
	if ok {
		k.lastPodData = pod
	} else if k.lastPodData != nil {
		pod = k.lastPodData
	}
	if pod != nil {
		containerName := getContainerNameFromPath(ev.Path)
		metadata := extractMetadataFromPod(pod, containerName)
		for k, v := range metadata {
			ev.Data["kubernetes."+k] = v
		}
	}
	return true
}

func extractMetadataFromPod(pod *v1.Pod, containerName string) map[string]interface{} {
	ret := make(map[string]interface{})
	pod.GetClusterName()
	ret["pod.labels"] = pod.Labels
	ret["pod.name"] = pod.Name
	ret["pod.namespace"] = pod.Namespace
	ret["pod.resourceVersion"] = pod.ResourceVersion
	ret["pod.UID"] = string(pod.UID)
	ret["pod.nodeName"] = pod.Spec.NodeName
	ret["pod.nodeSelector"] = pod.Spec.NodeSelector
	ret["pod.serviceAccountName"] = pod.Spec.ServiceAccountName
	ret["pod.subdomain"] = pod.Spec.Subdomain
	ret["pod.annotations"] = pod.Annotations

	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			ret["container.args"] = container.Args
			ret["container.command"] = container.Command
			ret["container.name"] = container.Name
			ret["container.env"] = container.Env
			ret["container.image"] = container.Image
			ret["container.ports"] = container.Ports
			ret["container.VolumeMounts"] = container.VolumeMounts
			ret["container.workingDir"] = container.WorkingDir
			ret["container.resources"] = container.Resources
		}
	}
	return ret
}

func getContainerNameFromPath(filePath string) string {
	r := regexp.MustCompile("/var/log/pods/(.*)/(?P<name>.*)/[0-9]+.log")
	match := r.FindStringSubmatch(filePath)
	if match == nil {
		return ""
	}
	if len(match) < 3 {
		return ""
	}
	return match[2]
}
