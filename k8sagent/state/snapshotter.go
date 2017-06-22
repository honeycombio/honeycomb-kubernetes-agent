package state

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

// Snapshotter recieves and fulfills queries to the Kubernetes API
// server.
type Snapshotter interface {
	Snapshot() error
}

// ----------------------------------------------------------------------------
// Implementation.
// ----------------------------------------------------------------------------

type snapshotterImpl struct {
	client        *kubernetes.Clientset
	name          string
	labelSelector string
	record        Record
}

// NewSnapshotter creates a new state snapshotter.
func NewSnapshotter(
	record Record, kubeconfig string, name string, labelSelector string,
) (Snapshotter, error) {
	kubeClientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, err
	}

	return &snapshotterImpl{
		client:        clientset,
		name:          name,
		labelSelector: labelSelector,
		record:        record,
	}, nil
}

// Labels queries the API server for list of pods that have some set of labels.
func (s *snapshotterImpl) Snapshot() error {
	nodes, err := s.client.Nodes().List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		if node.Name == s.name {
			pods, err := s.client.Pods(node.GetNamespace()).List(
				v1.ListOptions{
					LabelSelector: s.labelSelector,
				})
			if err != nil {
				return err
			}
			s.record.Replace(pods)

			return nil
		}
	}

	return fmt.Errorf("Could not retrieve list of pods")
}
