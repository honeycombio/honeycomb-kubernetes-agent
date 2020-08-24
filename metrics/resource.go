package metrics

import (
	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	"time"
)

type Resource struct {
	Type      string
	Name      string
	Labels    map[string]string
	Status    map[string]string
	Timestamp time.Time
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

	addAdditionalLabels(labels, node.Labels, s.PodRef.UID, metadata)

	status := metadata.GetStatusForPod(s.PodRef.UID)

	return &Resource{
		Type:      "pod",
		Name:      s.PodRef.Name,
		Labels:    labels,
		Status:    status,
		Timestamp: s.CPU.Time.Time,
	}
}

func getContainerResource(pod *Resource, s stats.ContainerStats, metadata *Metadata) (*Resource, error) {
	labels := map[string]string{
		LabelContainerName: s.Name,
	}

	addAdditionalLabels(labels, pod.Labels, pod.Labels[LabelPodUid], metadata)

	status := metadata.GetStatusForContainer(pod.Labels[LabelPodUid], s.Name)

	return &Resource{
		Type:      "container",
		Name:      s.Name,
		Labels:    labels,
		Status:    status,
		Timestamp: s.CPU.Time.Time,
	}, nil
}

func getVolumeResource(pod *Resource, s stats.VolumeStats, metadata *Metadata) *Resource {
	labels := map[string]string{
		LabelVolumeName: s.Name,
	}

	addAdditionalLabels(labels, pod.Labels, pod.Labels[LabelPodUid], metadata)

	return &Resource{
		Type:      "volume",
		Name:      s.Name,
		Labels:    labels,
		Timestamp: s.Time.Time,
	}
}

func addAdditionalLabels(labels map[string]string, resLabels map[string]string, uid string, metadata *Metadata) {
	// add resource labels
	for k, v := range resLabels {
		labels[k] = v
	}

	podLabels := metadata.GetLabelsForPod(uid)
	for k, v := range podLabels {
		labels[PrefixLabel+k] = v
	}
}
