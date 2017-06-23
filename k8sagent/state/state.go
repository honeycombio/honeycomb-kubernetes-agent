package state

import "k8s.io/client-go/pkg/api/v1"

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

type Record interface {
	Replace(string, *v1.PodList)
	Pods(string) (*v1.PodList, bool)
}

// ----------------------------------------------------------------------------
// Implementation.
// ----------------------------------------------------------------------------

type recordImpl struct {
	matchingPods map[string]*v1.PodList
}

func NewRecord() *recordImpl {
	return &recordImpl{
		matchingPods: make(map[string]*v1.PodList),
	}
}

func (r *recordImpl) Replace(labelSelector string, pl *v1.PodList) {
	// TODO: Lock me.
	r.matchingPods[labelSelector] = pl
}

func (r *recordImpl) Pods(labelSelector string) (*v1.PodList, bool) {
	pl, ok := r.matchingPods[labelSelector]
	return pl, ok
}
