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

func (k *KubernetesMetadataProcessor) Process(data map[string]interface{}) {
	podData, ok := k.PodGetter.Get(k.UID)
	if ok {
		data["kubernetes.pod"] = podData
		k.lastPodData = podData
	} else if k.lastPodData != nil {
		data["kubernetes.pod"] = k.lastPodData
	}
	// TODO add container metadata too
}
