package metrics

import (
	"k8s.io/apimachinery/pkg/types"
	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	"time"
)

type Resource struct {
	Type        string
	Name        string
	Labels      map[string]string
	Timestamp   time.Time
	Status      map[string]string
	PodMetadata *PodMetadata
}

func getNodeResource(s stats.NodeStats) *Resource {
	labels := map[string]string{
		LabelNodeName: s.NodeName,
	}

	return &Resource{
		Type:      "node",
		Name:      s.NodeName,
		Labels:    labels,
		Timestamp: s.CPU.Time.Time,
	}
}

func getPodResource(node *Resource, s stats.PodStats, metadata *Metadata) *Resource {
	labels := map[string]string{
		LabelNamespaceName: s.PodRef.Namespace,
		LabelPodUid:        s.PodRef.UID,
		LabelPodName:       s.PodRef.Name,
	}

	podMetadata, err := metadata.GetPodMetadataByUid(types.UID(s.PodRef.UID))
	if err != nil {
		return &Resource{
			Type:      "pod",
			Name:      s.PodRef.Name,
			Labels:    labels,
			Timestamp: s.CPU.Time.Time,
		}
	}

	status := podMetadata.GetStatus()
	addAdditionalLabels(labels, node.Labels, podMetadata)

	return &Resource{
		Type:        "pod",
		Name:        s.PodRef.Name,
		Labels:      labels,
		Timestamp:   s.CPU.Time.Time,
		Status:      status,
		PodMetadata: podMetadata,
	}

}

func getContainerResource(pod *Resource, s stats.ContainerStats) (*Resource, error) {
	labels := map[string]string{
		LabelContainerName: s.Name,
	}

	addAdditionalLabels(labels, pod.Labels, pod.PodMetadata)

	status := pod.PodMetadata.GetStatusForContainer(s.Name)

	return &Resource{
		Type:        "container",
		Name:        s.Name,
		Labels:      labels,
		Timestamp:   s.CPU.Time.Time,
		Status:      status,
		PodMetadata: pod.PodMetadata,
	}, nil
}

func getVolumeResource(pod *Resource, s stats.VolumeStats) *Resource {
	labels := map[string]string{
		LabelVolumeName: s.Name,
	}

	addAdditionalLabels(labels, pod.Labels, pod.PodMetadata)

	return &Resource{
		Type:        "volume",
		Name:        s.Name,
		Labels:      labels,
		Timestamp:   s.Time.Time,
		PodMetadata: pod.PodMetadata,
	}
}

func addAdditionalLabels(labels map[string]string, resLabels map[string]string, podMetadata *PodMetadata) {
	// add resource labels
	for k, v := range resLabels {
		labels[k] = v
	}

	podLabels := podMetadata.GetLabels()
	for k, v := range podLabels {
		labels[PrefixLabel+k] = v
	}
}
