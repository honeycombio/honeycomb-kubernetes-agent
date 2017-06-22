package state

import "k8s.io/client-go/pkg/api/v1"

// ----------------------------------------------------------------------------
// Core interfaces.
// ----------------------------------------------------------------------------

type Record interface {
	Replace(pl *v1.PodList)
	Pods() *v1.PodList
}

// ----------------------------------------------------------------------------
// Implementation.
// ----------------------------------------------------------------------------

type recordImpl struct {
	matchingPods *v1.PodList
}

func NewRecord() *recordImpl {
	return &recordImpl{}
}

func (r *recordImpl) Replace(pl *v1.PodList) {
	// TODO: Lock me.
	r.matchingPods = pl
}

func (r *recordImpl) Pods() *v1.PodList {
	// TODO: Lock me.
	return r.matchingPods
}
