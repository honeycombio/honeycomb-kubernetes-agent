package state

import (
	"fmt"

	api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

// Snapshotter recieves and fulfills queries to the Kubernetes API
// server.
type Snapshotter interface {
	// TODO: Extend this to monitor for updates.
	Snapshot() error
}

// ----------------------------------------------------------------------------
// Implementation.
// ----------------------------------------------------------------------------

type snapshotterImpl struct {
	labelSelectors []string
	client         *kubernetes.Clientset
	host           *v1.Node
	record         Record
}

// NewSnapshotter creates a new state snapshotter.
func NewSnapshotter(labelSelectors []string,
	record Record, name string,
) (Snapshotter, error) {
	// Get clientset to query API server.
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, err
	}

	// Try to resolve local node with `name`.
	// TODO: Consider using `FieldSelector` to select this node.
	nodes, err := clientset.Nodes().List(api_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var host *v1.Node
	for _, node := range nodes.Items {
		if node.Name == name {
			host = &node
		}
	}

	if host == nil {
		return nil, fmt.Errorf("Could not find node with name '%s'", name)
	}

	// Success.
	return &snapshotterImpl{
		labelSelectors: labelSelectors,
		client:         clientset,
		host:           host,
		record:         record,
	}, nil
}

// Snapshot queries the API server for list of pods that have some set of labels.
func (s *snapshotterImpl) Snapshot() error {
	for _, labelSelector := range s.labelSelectors {
		// NOTE: The `""` here denotes that we should search all
		// namespaces for pods.
		pods, err := s.client.Pods("").List(
			api_v1.ListOptions{
				LabelSelector: labelSelector,
				FieldSelector: fmt.Sprintf("spec.nodeName=%s", s.host.Name),
			})
		if err != nil {
			return err
		}

		s.record.Replace(labelSelector, pods)
	}

	return nil
}
