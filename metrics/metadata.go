package metrics

import (
	"errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

type OmitLabel string

type Metadata struct {
	PodsMetadata *v1.PodList
	OmitLabels   []OmitLabel
	logger       *logrus.Logger
}

func NewMetadata(podsMetadata *v1.PodList, omitLabels []OmitLabel, logger *logrus.Logger) *Metadata {
	return &Metadata{
		PodsMetadata: podsMetadata,
		OmitLabels:   omitLabels,
		logger:       logger,
	}
}

func (m *Metadata) GetLabelsForPod(uid string) map[string]string {
	pod, err := m.getPodByUid(types.UID(uid))
	if err != nil {
		return nil
	}

	labels := make(map[string]string)
	for k, v := range pod.Labels {
		keep := true
		for _, o := range m.OmitLabels {
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

func (m *Metadata) GetStatusForPod(uid string) map[string]string {
	pod, err := m.getPodByUid(types.UID(uid))
	if err != nil {
		return nil
	}

	status := map[string]string{
		StatusPodMessage: pod.Status.Message,
		StatusPodPhase:   string(pod.Status.Phase),
		StatusPodReason:  pod.Status.Reason,
	}

	return status
}

func (m *Metadata) GetStatusForContainer(uid string, name string) map[string]string {
	pod, err := m.getPodByUid(types.UID(uid))
	if err != nil {
		return nil
	}

	var s v1.ContainerStatus
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name {
			s = cs
			break
		}
	}
	if s.ContainerID == "" {
		m.logger.WithFields(logrus.Fields{
			"podUid":        uid,
			"containerName": name,
		}).Error("Metadata: Container not found")
		return nil
	}

	status := map[string]string{
		LabelContainerId:        s.Name,
		StatusContainerReady:    strconv.FormatBool(s.Ready),
		StatusContainerRestarts: string(s.RestartCount),
	}

	if s.State.Running != nil {
		status[StatusContainerState] = "running"
	} else if s.State.Terminated != nil {
		status[StatusContainerState] = "terminated"
		status[StatusContainerExitCode] = string(s.State.Terminated.ExitCode)
		status[StatusContainerMessage] = s.State.Terminated.Message
		status[StatusContainerReason] = s.State.Terminated.Reason
	} else if s.State.Waiting != nil {
		status[StatusContainerState] = "waiting"
		status[StatusContainerMessage] = s.State.Waiting.Message
		status[StatusContainerReason] = s.State.Waiting.Reason
	} else {
		status[StatusContainerState] = "waiting"
	}

	return status
}

func (m *Metadata) getPodByUid(uid types.UID) (v1.Pod, error) {
	for _, p := range m.PodsMetadata.Items {
		if p.UID == uid {
			return p, nil
		}
	}

	m.logger.WithFields(logrus.Fields{
		"podUid": uid,
	}).Error("Metadata: Pod not found")
	return v1.Pod{}, errors.New("pod not found")
}
