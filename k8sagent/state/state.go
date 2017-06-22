package state

import "k8s.io/client-go/pkg/api/v1"

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

type Record interface {
	Replace(dataset string, pl *v1.PodList)
	Pods(dataset string) (*v1.PodList, bool)
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

func (r *recordImpl) Replace(dataset string, pl *v1.PodList) {
	// TODO: Lock me.
	r.matchingPods[dataset] = pl
}

func (r *recordImpl) Pods(dataset string) (*v1.PodList, bool) {
	// TODO: Lock me.
	pl, ok := r.matchingPods[dataset]
	return pl, ok
}
