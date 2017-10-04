// Package k8sagent contains machinery for getting pod data from the Kubernetes
// API.
package k8sagent

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/types"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher interface {
	Get(types.UID) (*v1.Pod, bool)
	Pods() chan *v1.Pod
}

type PodWatcherImpl struct {
	watchMap  map[types.UID]*v1.Pod
	watchChan chan *v1.Pod
	client    corev1.PodsGetter
	sync.RWMutex
}

func NewPodWatcher(namespace string, labelSelector string, fieldSelector string, client corev1.PodsGetter) *PodWatcherImpl {
	w := &PodWatcherImpl{
		client:    client,
		watchMap:  make(map[types.UID]*v1.Pod),
		watchChan: make(chan *v1.Pod, 100),
	}

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    w.onAdd,
		UpdateFunc: w.onUpdate,
		DeleteFunc: w.onDelete,
	}

	go runInformer(
		namespace,
		labelSelector,
		fieldSelector,
		client,
		handlers)
	return w
}

func (w *PodWatcherImpl) Pods() chan *v1.Pod {
	return w.watchChan
}

func (w *PodWatcherImpl) onAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		logrus.Error("Watcher got new object of unexpected type")
		return
	}

	w.Lock()
	defer w.Unlock()
	_, ok = w.watchMap[pod.UID]
	w.watchMap[pod.UID] = pod
	if !ok {
		w.watchChan <- pod
	}

}

func (w *PodWatcherImpl) onUpdate(oldObj interface{}, newObj interface{}) {
	pod, ok := newObj.(*v1.Pod)
	if !ok {
		logrus.Error("Watcher got new object of unexpected type")
		return
	}
	w.Lock()
	defer w.Unlock()
	_, ok = w.watchMap[pod.UID]
	w.watchMap[pod.UID] = pod
	if !ok {
		w.watchChan <- pod
	}

}

func (w *PodWatcherImpl) onDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	// TODO check for deleted type
	if !ok {
		logrus.Error("Watcher got new object of unexpected type")
		return
	}
	w.Lock()
	defer w.Unlock()
	delete(w.watchMap, pod.UID)
}

func (w *PodWatcherImpl) Get(uid types.UID) (*v1.Pod, bool) {
	w.RLock()
	defer w.RUnlock()
	v, ok := w.watchMap[uid]
	return v, ok
}

func runInformer(
	namespace string,
	labelSelector string,
	fieldSelector string,
	client corev1.PodsGetter,
	handler cache.ResourceEventHandlerFuncs,
) {
	listOptions := v1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}
	listFunc := func(options api.ListOptions) (runtime.Object, error) {
		return client.Pods(namespace).List(listOptions)
	}
	watchFunc := func(options api.ListOptions) (watch.Interface, error) {
		return client.Pods(namespace).Watch(listOptions)
	}

	watchlist := &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
	resyncPeriod := time.Minute
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		resyncPeriod,
		handler,
	)

	logrus.WithFields(logrus.Fields{
		"labelSelector": labelSelector,
		"namespace":     namespace,
		"fieldSelector": fieldSelector,
	}).Debug("Starting informer")
	go controller.Run(wait.NeverStop)
}
