// Package podtailer contains machinery for tailing the logs of a *set* of pods
// matching a labelSelector.
package podtailer

import (
	"fmt"
	"regexp"
	"sync"

	"k8s.io/client-go/pkg/api/v1"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// PodSetTailer is responsible for watching for all pods that match the
// criteria defined by config, and managing tailers for each pod.
type PodSetTailer struct {
	config        *config.WatcherConfig
	nodeSelector  string
	transmitter   transmission.Transmitter
	stateRecorder tailer.StateRecorder
	kubeClient    corev1.PodsGetter
	stop          chan bool
	wg            sync.WaitGroup
}

func NewPodSetTailer(
	config *config.WatcherConfig,
	nodeSelector string,
	transmitter transmission.Transmitter,
	stateRecorder tailer.StateRecorder,
	kubeClient corev1.PodsGetter,
) *PodSetTailer {
	return &PodSetTailer{
		config:        config,
		nodeSelector:  nodeSelector,
		transmitter:   transmitter,
		stateRecorder: stateRecorder,
		kubeClient:    kubeClient,
		stop:          make(chan bool),
	}
}

func (pt *PodSetTailer) run() {
	defer pt.wg.Done()
	labelSelector := *pt.config.LabelSelector
	// Exclude the agent's own logs from being watched
	if labelSelector == "" {
		labelSelector = "k8s-app!=honeycomb-agent"
	} else {
		labelSelector = labelSelector + ",k8s-app!=honeycomb-agent"
	}

	podWatcher := k8sagent.NewPodWatcher(
		pt.config.Namespace,
		labelSelector,
		pt.nodeSelector,
		pt.kubeClient)

loop:
	for {
		select {
		case pod := <-podWatcher.Pods():
			watcher, err := pt.watcherForPod(pod, pt.config.ContainerName, podWatcher)
			if err != nil {
				// This shouldn't happen, since we check for configuration errors
				// before actually setting up the watcher
				logrus.WithError(err).Error("Error setting up watcher")
				continue loop
			}
			watcher.Start()
			defer watcher.Stop()
		case <-pt.stop:
			break loop
		}
	}
}

func (pt *PodSetTailer) Start() {
	pt.wg.Add(1)
	go pt.run()
}

func (pt *PodSetTailer) Stop() {
	pt.stop <- true
	pt.wg.Wait()
}

func (pt *PodSetTailer) watcherForPod(pod *v1.Pod, containerName string, podWatcher k8sagent.PodWatcher) (*tailer.PathWatcher, error) {
	path := fmt.Sprintf("/var/log/pods/%s", pod.UID)
	pattern := fmt.Sprintf("%s/*", path)
	var filterFunc func(fileName string) bool

	if containerName != "" {
		// only watch logs for containers matching the given name, if
		// one is specified
		re := fmt.Sprintf("^%s/%s_[0-9]*\\.log", path, regexp.QuoteMeta(containerName))
		filterFunc = func(fileName string) bool {
			ok, _ := regexp.Match(re, []byte(fileName))
			return ok
		}
	}

	k8sMetadataProcessor := &processors.KubernetesMetadataProcessor{
		PodGetter: podWatcher,
		UID:       pod.UID}
	handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
		pt.config,
		&unwrappers.DockerJSONLogUnwrapper{},
		pt.transmitter,
		k8sMetadataProcessor)
	if err != nil {
		// This shouldn't happen, since we check for configuration errors
		// before actually setting up the watcher
		logrus.WithError(err).Error("Error setting up watcher")
		return nil, err
	}

	watcher := tailer.NewPathWatcher(pattern, filterFunc, handlerFactory, pt.stateRecorder)
	return watcher, nil
}
