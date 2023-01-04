package metrics

import (
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type OmitLabel string

type Metadata struct {
	PodsMetadata      *v1.PodList
	NodesMetadata     *v1.NodeList
	OmitLabels        []OmitLabel
	IncludeNodeLabels bool
	IncludeNodeInfo   bool
}

func NewMetadata(podsMetadata *v1.PodList, nodesMetadata *v1.NodeList, omitLabels []OmitLabel, includeNodeLabels bool) *Metadata {
	return &Metadata{
		PodsMetadata:      podsMetadata,
		NodesMetadata:     nodesMetadata,
		OmitLabels:        omitLabels,
		IncludeNodeLabels: includeNodeLabels,
	}
}

func (m *Metadata) GetNodeMetadataByName(name string) (*NodeMetadata, error) {
	for _, n := range m.NodesMetadata.Items {
		if n.ObjectMeta.Name == name {
			return &NodeMetadata{m, &n}, nil
		}
	}

	logrus.WithFields(logrus.Fields{
		"nodeName": name,
	}).Error("Metadata: Node not found")
	return &NodeMetadata{}, errors.New("node not found")
}

func (m *Metadata) GetPodMetadataByUid(uid types.UID) (*PodMetadata, error) {
	for _, p := range m.PodsMetadata.Items {
		if p.UID == uid {
			return &PodMetadata{m, &p}, nil
		}
	}

	logrus.WithFields(logrus.Fields{
		"podUid": uid,
	}).Error("Metadata: Pod not found")
	return &PodMetadata{}, errors.New("pod not found")
}

type NodeMetadata struct {
	Metadata *Metadata
	Node     *v1.Node
}

func (n *NodeMetadata) GetLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range n.Node.Labels {
		keep := true
		for _, o := range n.Metadata.OmitLabels {
			if o == OmitLabel(k) {
				keep = false
				break
			}
		}
		if keep {
			labels[k] = v
		}
	}
	return labels
}

func (n *NodeMetadata) GetNodeInfo() map[string]string {
	return map[string]string{
		"architecture":            n.Node.Status.NodeInfo.Architecture,
		"containerRuntimeVersion": n.Node.Status.NodeInfo.ContainerRuntimeVersion,
		"kernelVersion":           n.Node.Status.NodeInfo.KernelVersion,
		"kubeProxyVersion":        n.Node.Status.NodeInfo.KubeProxyVersion,
		"kubeletVersion":          n.Node.Status.NodeInfo.KubeletVersion,
		"operatingSystem":         n.Node.Status.NodeInfo.OperatingSystem,
		"osImage":                 n.Node.Status.NodeInfo.OSImage,
	}
}

type PodMetadata struct {
	Metadata *Metadata
	Pod      *v1.Pod
}

func (p *PodMetadata) GetLabels() map[string]string {
	labels := make(map[string]string)
	for k, v := range p.Pod.Labels {
		keep := true
		for _, o := range p.Metadata.OmitLabels {
			if o == OmitLabel(k) {
				keep = false
				break
			}
		}
		if keep {
			labels[k] = v
		}
	}
	return labels
}

func (p *PodMetadata) GetCpuLimit() float64 {
	var limit float64
	for _, c := range p.Pod.Spec.Containers {
		val, err := strconv.ParseFloat(c.Resources.Limits.Cpu().AsDec().String(), 64)
		if err != nil || val == 0 {
			return 0
		}
		limit += val
	}
	return limit
}

func (p *PodMetadata) GetCpuLimitForContainer(name string) float64 {
	for _, c := range p.Pod.Spec.Containers {
		if c.Name == name {
			limit, _ := strconv.ParseFloat(c.Resources.Limits.Cpu().AsDec().String(), 64)
			return limit
		}
	}
	return 0
}

func (p *PodMetadata) GetMemoryLimit() float64 {
	var limit float64
	for _, c := range p.Pod.Spec.Containers {
		val, err := strconv.ParseFloat(c.Resources.Limits.Memory().AsDec().String(), 64)
		if err != nil || val == 0 {
			return 0
		}
		limit += val
	}
	return limit
}

func (p *PodMetadata) GetMemoryLimitForContainer(name string) float64 {
	for _, c := range p.Pod.Spec.Containers {
		if c.Name == name {
			limit, _ := strconv.ParseFloat(c.Resources.Limits.Memory().AsDec().String(), 64)
			return limit
		}
	}
	return 0
}

func (p *PodMetadata) GetStatus() map[string]interface{} {

	status := map[string]interface{}{
		StatusMessage: p.Pod.Status.Message,
		StatusPhase:   string(p.Pod.Status.Phase),
		StatusReason:  p.Pod.Status.Reason,
	}

	// Pod restarts is the max restart count from all containers
	restarts := int32(0)
	for _, cs := range p.Pod.Status.ContainerStatuses {
		if restarts < cs.RestartCount {
			restarts = cs.RestartCount
		}
	}
	status[StatusRestartCount] = restarts

	return status
}

func (p *PodMetadata) GetStatusForContainer(name string) map[string]interface{} {

	var s v1.ContainerStatus
	for _, cs := range p.Pod.Status.ContainerStatuses {
		if cs.Name == name {
			s = cs
			break
		}
	}
	if s.ContainerID == "" {
		logrus.WithFields(logrus.Fields{
			"podName":       p.Pod.Name,
			"containerName": name,
		}).Error("Metadata: Container not found")
		return nil
	}

	status := map[string]interface{}{
		LabelContainerId:   s.Name,
		StatusReady:        s.Ready,
		StatusRestartCount: s.RestartCount,
	}

	if s.State.Running != nil {
		status[StatusState] = "running"
	} else if s.State.Terminated != nil {
		status[StatusState] = "terminated"
		status[StatusExitCode] = s.State.Terminated.ExitCode
		status[StatusMessage] = s.State.Terminated.Message
		status[StatusReason] = s.State.Terminated.Reason
	} else if s.State.Waiting != nil {
		status[StatusState] = "waiting"
		status[StatusMessage] = s.State.Waiting.Message
		status[StatusReason] = s.State.Waiting.Reason
	} else {
		status[StatusState] = "waiting"
	}

	return status
}
