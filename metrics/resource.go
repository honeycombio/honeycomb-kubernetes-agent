package metrics

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type Resource struct {
	Type        string
	Name        string
	Labels      map[string]string
	Timestamp   time.Time `json:"time,omitempty"`
	Status      map[string]interface{}
	PodMetadata *PodMetadata
}

func getNodeResource(s stats.NodeStats) *Resource {
	labels := map[string]string{
		LabelNodeName: s.NodeName,
	}

	resource := &Resource{
		Type:   "node",
		Name:   s.NodeName,
		Labels: labels,
	}
	if s.CPU != nil {
		resource = &Resource{
			Type:      "node",
			Name:      s.NodeName,
			Labels:    labels,
			Timestamp: s.CPU.Time.Time,
		}
	}
	return resource
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
			Type:   "pod",
			Name:   s.PodRef.Name,
			Labels: labels,
		}
	}

	status := podMetadata.GetStatus()
	addAdditionalLabels(labels, node.Labels, podMetadata)

	resource := &Resource{
		Type:        "pod",
		Name:        s.PodRef.Name,
		Labels:      labels,
		Status:      status,
		PodMetadata: podMetadata,
	}

	if s.CPU != nil {
		resource = &Resource{
			Type:        "pod",
			Name:        s.PodRef.Name,
			Labels:      labels,
			Timestamp:   s.CPU.Time.Time,
			Status:      status,
			PodMetadata: podMetadata,
		}
	}
	return resource

}

func getContainerResource(pod *Resource, s stats.ContainerStats) (*Resource, error) {
	labels := map[string]string{
		LabelContainerName: s.Name,
	}

	addAdditionalLabels(labels, pod.Labels, pod.PodMetadata)

	status := pod.PodMetadata.GetStatusForContainer(s.Name)

	resource := &Resource{
		Type:        "container",
		Name:        s.Name,
		Labels:      labels,
		Status:      status,
		PodMetadata: pod.PodMetadata,
	}
	if s.CPU != nil {
		resource = &Resource{
			Type:        "container",
			Name:        s.Name,
			Labels:      labels,
			Timestamp:   s.CPU.Time.Time,
			Status:      status,
			PodMetadata: pod.PodMetadata,
		}
	}
	return resource, nil
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
