package state

import "k8s.io/client-go/pkg/api/v1"

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

type Record interface {
	Replace(index int, pl *v1.PodList)
	Pods(index int) (*v1.PodList, bool)
}

// ----------------------------------------------------------------------------
// Implementation.
// ----------------------------------------------------------------------------

type recordImpl struct {
	matchingPods []*v1.PodList
}

func NewRecord(length int) *recordImpl {
	return &recordImpl{
		matchingPods: make([]*v1.PodList, length),
	}
}

func (r *recordImpl) Replace(index int, pl *v1.PodList) {
	// TODO: Lock me.
	r.matchingPods[index] = pl
}

func (r *recordImpl) Pods(index int) (*v1.PodList, bool) {
	// TODO: Lock me.
	if index < 0 || index >= len(r.matchingPods) {
		return nil, false
	}
	pl := r.matchingPods[index]
	return pl, true
}
