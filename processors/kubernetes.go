package processors

import (
	"path"
	"regexp"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
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
	r := regexp.MustCompile("(?P<name>.*)_[0-9]+.log")
	filename := path.Base(filePath)
	match := r.FindStringSubmatch(filename)
	if match == nil {
		return ""
	}
	if len(match) < 2 {
		return ""
	}
	return match[1]
}
