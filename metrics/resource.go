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
	Timestamp   time.Time
	Status      map[string]interface{}
	PodMetadata *PodMetadata
}

func getNodeResource(s stats.NodeStats, metadata *Metadata) *Resource {
	labels := map[string]string{
		LabelNodeName: s.NodeName,
	}
	if metadata.IncludeNodeLabels {
		nodeMetadata, err := metadata.GetNodeMetadataByName(s.NodeName)
		if err == nil {
			for k, v := range nodeMetadata.GetLabels() {
				labels[PrefixLabel+k] = v
			}
		}
	}
	resource := &Resource{
		Type:   "node",
		Name:   s.NodeName,
		Labels: labels,
	}

	if s.CPU != nil {
		resource.Timestamp = s.CPU.Time.Time
	}

	return resource
}

func getPodResource(node *Resource, s stats.PodStats, metadata *Metadata) *Resource {
	labels := map[string]string{
		LabelNamespaceName: s.PodRef.Namespace,
		LabelPodUid:        s.PodRef.UID,
		LabelPodName:       s.PodRef.Name,
	}

	resource := &Resource{
		Type:   "pod",
		Name:   s.PodRef.Name,
		Labels: labels,
	}

	if s.CPU != nil {
		resource.Timestamp = s.CPU.Time.Time
	}

	podMetadata, err := metadata.GetPodMetadataByUid(types.UID(s.PodRef.UID))
	if err != nil {
		return resource
	}

	status := podMetadata.GetStatus()
	addAdditionalLabels(labels, node.Labels, podMetadata)

	resource.Status = status
	resource.PodMetadata = podMetadata

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
		resource.Timestamp = s.CPU.Time.Time
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
