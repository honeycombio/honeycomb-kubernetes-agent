// Package k8sagent contains machinery for getting pod data from the Kubernetes
// API.
package k8sagent

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	v1types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher interface {
	Get(types.UID) (*v1.Pod, bool)
	Pods() chan *v1.Pod
	DeletedPods() chan types.UID
}

type PodWatcherImpl struct {
	watchMap    map[types.UID]*v1.Pod
	watchChan   chan *v1.Pod
	unwatchChan chan types.UID
	client      corev1.PodsGetter
	sync.RWMutex
}

func NewPodWatcher(namespace string, labelSelector string, fieldSelector string, client corev1.PodsGetter) *PodWatcherImpl {
	w := &PodWatcherImpl{
		client:      client,
		watchMap:    make(map[types.UID]*v1.Pod),
		watchChan:   make(chan *v1.Pod, 100),
		unwatchChan: make(chan types.UID, 100),
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

func (w *PodWatcherImpl) DeletedPods() chan types.UID {
	return w.unwatchChan
}

func (w *PodWatcherImpl) onAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		logrus.Error("Watcher got new object of unexpected type")
		return
	}

	logrus.WithFields(logrus.Fields{
		"UID":    pod.UID,
		"Name":   pod.Name,
		"Action": "Add",
	}).Debug("Watcher got new pod")

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

	logrus.WithFields(logrus.Fields{
		"UID":    pod.UID,
		"Name":   pod.Name,
		"Action": "Update",
	}).Debug("Watcher got updated pod")

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

	logrus.WithFields(logrus.Fields{
		"UID":    pod.UID,
		"Name":   pod.Name,
		"Action": "Delete",
	}).Debug("Watcher got pod")

	w.Lock()
	defer w.Unlock()
	delete(w.watchMap, pod.UID)
	w.unwatchChan <- pod.UID
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
	listOptions := v1types.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}
	listFunc := func(options v1types.ListOptions) (runtime.Object, error) {
		return client.Pods(namespace).List(context.Background(), listOptions)
	}
	watchFunc := func(options v1types.ListOptions) (watch.Interface, error) {
		return client.Pods(namespace).Watch(context.Background(), listOptions)
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
