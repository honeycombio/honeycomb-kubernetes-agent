// Package podtailer contains machinery for tailing the logs of a *set* of pods
// matching a labelSelector.
package podtailer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const logsBasePath = "/var/log"
const maxPodWatcherRetries = 3

// PodSetTailer is responsible for watching for all pods that match the
// criteria defined by config, and managing tailers for each pod.
type PodSetTailer struct {
	config                 *config.WatcherConfig
	nodeSelector           string
	transmitter            transmission.Transmitter
	stateRecorder          tailer.StateRecorder
	kubeClient             corev1.PodsGetter
	stop                   chan bool
	wg                     sync.WaitGroup
	legacyLogPaths         bool
	additionalFieldsGlobal map[string]interface{}
}

func NewPodSetTailer(
	config *config.WatcherConfig,
	nodeSelector string,
	transmitter transmission.Transmitter,
	stateRecorder tailer.StateRecorder,
	kubeClient corev1.PodsGetter,
	legacyLogPaths bool,
	additionalFieldsGlobal map[string]interface{},
) *PodSetTailer {
	return &PodSetTailer{
		config:                 config,
		nodeSelector:           nodeSelector,
		transmitter:            transmitter,
		stateRecorder:          stateRecorder,
		kubeClient:             kubeClient,
		stop:                   make(chan bool),
		legacyLogPaths:         legacyLogPaths,
		additionalFieldsGlobal: additionalFieldsGlobal,
	}
}

func (pt *PodSetTailer) run() {
	defer pt.wg.Done()
	labelSelector := *pt.config.LabelSelector
	// Exclude the agent's own logs from being watched
	// This is quasi-hack. Though we recommend the use of the `app: honeycomb-agent` label, this can't be enforced.
	if labelSelector == "" {
		labelSelector = "app!=honeycomb-agent"
	} else {
		labelSelector = labelSelector + ",app!=honeycomb-agent"
	}

	podWatcher := k8sagent.NewPodWatcher(
		pt.config.Namespace,
		labelSelector,
		pt.nodeSelector,
		pt.kubeClient)

	watcherMap := make(map[types.UID]*tailer.PathWatcher)

loop:
	for {
		select {
		case pod := <-podWatcher.Pods():
			watcher, err := pt.watcherForPod(pod, pt.config.ContainerName, podWatcher)
			if err != nil {

				// sometimes the pod isn't ready - for example, its log directory might not be
				// on disk yet - back off and retry on incremental intervals before giving up
				for retries := 1; retries <= maxPodWatcherRetries; retries++ {
					watcher, err = pt.watcherForPod(pod, pt.config.ContainerName, podWatcher)
					if err == nil {
						break
					}
					if retries == maxPodWatcherRetries {
						logrus.WithError(err).WithFields(logrus.Fields{
							"name":      pod.Name,
							"uid":       pod.UID,
							"namespace": pod.Namespace,
						}).Error("Error setting up watcher, giving up on this pod")
						continue loop
					}
					logrus.WithError(err).WithFields(logrus.Fields{
						"name":           pod.Name,
						"uid":            pod.UID,
						"namespace":      pod.Namespace,
						"retry_attempts": retries,
					}).Warn("got error setting up watcher, retrying")
					time.Sleep(time.Second * time.Duration(retries) * time.Duration(retries))
				}
			}
			if watcher != nil {
				logrus.WithFields(logrus.Fields{
					"labelselector": labelSelector,
					"name":          pod.Name,
					"uid":           pod.UID,
					"namespace":     pod.Namespace,
				}).Info("starting watcher for pod")
				watcher.Start()
				watcherMap[pod.UID] = watcher
			} else {
				// should never get here, but just in case...
				logrus.WithFields(logrus.Fields{
					"name":      pod.Name,
					"uid":       pod.UID,
					"namespace": pod.Namespace,
				}).Warn("got nil watcher for pod")
			}
		case deletedPodUID := <-podWatcher.DeletedPods():
			if watcher, ok := watcherMap[deletedPodUID]; ok {
				logrus.WithFields(logrus.Fields{
					"uid": deletedPodUID,
				}).Info("pod deleted, stopping watcher")
				watcher.Stop()
				delete(watcherMap, deletedPodUID)
			}
		case <-pt.stop:
			break loop
		}
	}

	for key := range watcherMap {
		watcherMap[key].Stop()
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

// If you change the log pattern, make sure to check that the filter pattern still works
// (in determineFilterFunc)
func determineLogPattern(pod *v1.Pod, basePath string, legacyLogPaths bool) (string, error) {
	if basePath == "" {
		basePath = logsBasePath
	}

	// Welcome to "What log pattern is it, anyway?" The show where the log paths are made up and nothing matters.
	// Here we guess where the pod logs are, since k8s changes them every few versions, but doesn't provide an API
	// for asking the cluster for the path.

	// Legacy pattern was:
	// /var/log/containers/<pod_name>_<pod_namespace>_<container_name>-<container_id>.log`
	// For now, this is still supported on newer k8s clusters with a symlink
	if legacyLogPaths {
		return filepath.Join(basePath, "containers", fmt.Sprintf("%s_%s_*.log", pod.Name, pod.Namespace)), nil
	}
	// Added in https://github.com/kubernetes/kubernetes/pull/74441 - later k8s instances use
	// /var/log/pods/NAMESPACE_NAME_UID/CONTAINER/INSTANCE.log
	// If it exists, assume this log patern
	namespaceNameUIDPath := filepath.Join(basePath, "pods", fmt.Sprintf("%s_%s_%s", pod.Namespace, pod.Name, pod.UID))
	if _, err := os.Stat(namespaceNameUIDPath); err == nil {
		return filepath.Join(namespaceNameUIDPath, "*", "*"), nil
	}

	// Critical pods seem to all use this config hash for their log directory
	// instead of the pod UID. Use the hash if it exists
	if hash, ok := pod.Annotations["kubernetes.io/config.hash"]; ok {
		// Some newer system pods use NAMESPACE_NAME_HASH, so we have to support that too
		namespaceNameHashPath := filepath.Join(basePath, "pods", fmt.Sprintf("%s_%s_%s", pod.Namespace, pod.Name, hash))
		if _, err := os.Stat(namespaceNameHashPath); err == nil {
			return filepath.Join(namespaceNameHashPath, "*", "*"), nil
		}
		hpath := filepath.Join(basePath, "pods", hash)
		if _, err := os.Stat(hpath); err == nil {
			logrus.WithFields(logrus.Fields{
				"PodName": pod.Name,
				"UID":     pod.UID,
				"Hash":    hash,
			}).Info("Critical pod detected, using config.hash for log dir")

			// Now look for directories. Some newer k8s use /var/log/pods/<hash>/<podname>/NNN.log
			f, err := os.Open(hpath)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"PodName": pod.Name,
					"UID":     pod.UID,
					"Path":    hpath,
				}).WithError(err).Warn("failed to open pod log directory")
				return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
			}

			files, err := f.Readdirnames(-1)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"PodName": pod.Name,
					"UID":     pod.UID,
					"Path":    hpath,
				}).WithError(err).Warn("failed to read pod log directory")
				return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
			}
			// it's possible that the pod log directory exists, but no files or directories
			// exist so we can't know what the pattern should be yet
			if len(files) == 0 {
				logrus.WithFields(logrus.Fields{
					"PodName": pod.Name,
					"UID":     pod.UID,
					"Path":    hpath,
				}).Info("no files in log path yet, could not determine log pattern")
				return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
			}
			for _, f := range files {
				if s, err := os.Stat(filepath.Join(hpath, f)); err == nil {
					if s.IsDir() {
						return filepath.Join(hpath, "*", "*"), nil
					}
				}
			}

			// older k8s use /var/log/pods/hash/NNN.log
			return filepath.Join(hpath, "*"), nil
		}
	}
	upath := filepath.Join(basePath, "pods", string(pod.UID))
	if _, err := os.Stat(upath); err == nil {
		// is this a k8s 1.10 path?
		// pattern is /var/log/pods/<podUID>/<containerName>/<instance#>.log
		// So we want to check: are there directories in path or just files?
		f, err := os.Open(upath)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"PodName": pod.Name,
				"UID":     pod.UID,
				"Path":    upath,
			}).WithError(err).Warn("failed to open pod log directory")
			// should be unlikely, since we just stat'd the dir, but it will happen
			return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
		}

		files, err := f.Readdirnames(-1)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"PodName": pod.Name,
				"UID":     pod.UID,
				"Path":    upath,
			}).WithError(err).Warn("failed to read pod log directory")
			return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
		}
		// it's possible that the pod log directory exists, but no files or directories
		// exist so we can't know what the pattern should be yet
		if len(files) == 0 {
			logrus.WithFields(logrus.Fields{
				"PodName": pod.Name,
				"UID":     pod.UID,
				"Path":    upath,
			}).Info("no files in log path yet, could not determine log pattern")
			return "", fmt.Errorf("Could not determine log path for pod %s", pod.UID)
		}
		for _, f := range files {
			if s, err := os.Stat(filepath.Join(upath, f)); err == nil {
				// if we find at least one directory in the path, assume k8s
				// 1.10 pattern
				if s.IsDir() {
					return filepath.Join(upath, "*", "*"), nil
				}
			}
		}

		// older pattern is
		// /var/log/pods/<podUID>/<containerName>_<instance#>.log
		return filepath.Join(upath, "*"), nil
	}
	return "", fmt.Errorf("Could not find specified log path for pod %s", pod.UID)
}

func determineFilterFunc(pod *v1.Pod, basePath string, containerName string, legacyLogPaths bool, excludePaths []string) func(fileName string) bool {
	// this is an exclude function that conforms to the filter specs -- it returns true
	// if the file should be *included*. By default, it's nil (which is treated as true)
	var exclusionFilter func(string) bool
	if len(excludePaths) > 0 {
		excludes := make([]string, 0)
		for _, exclude := range excludePaths {
			// don't try to evaluate patterns that we can't even parse
			if doublestar.ValidatePattern(exclude) {
				excludes = append(excludes, exclude)
				logrus.WithField("exclude", exclude).Debug("Excluding path")
			}
		}
		exclusionFilter = func(fileName string) bool {
			// check if we should exclude this path
			for _, exclude := range excludes {
				shouldExclude, _ := doublestar.Match(exclude, fileName)
				if shouldExclude {
					return false // false means exclude
				}
			}
			return true
		}
	}

	if containerName == "" {
		if len(excludePaths) == 0 {
			logrus.Debug("No container name or exclude paths specified, no filter function needed")
			return nil
		}
		return exclusionFilter
	}

	if legacyLogPaths {
		re := fmt.Sprintf(
			"^%s/containers/%s_%s_%s-.+\\.log",
			basePath,
			pod.Name,
			pod.Namespace,
			containerName,
		)
		logrus.WithFields(logrus.Fields{
			"regex": re,
		}).Debug("Container filter function")

		return func(fileName string) bool {
			// first check if we should exclude this path
			// If exclusionFilter says no, we're going to say no.
			if exclusionFilter != nil && !exclusionFilter(fileName) {
				return false
			}
			ok, _ := regexp.Match(re, []byte(fileName))
			return ok
		}
	}

	uid := string(pod.UID)
	if hash, ok := pod.Annotations["kubernetes.io/config.hash"]; ok {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Debug("Using hash for uID")
		uid = hash
	}

	// HACK: try the https://github.com/kubernetes/kubernetes/pull/74441 log pattern
	re1 := fmt.Sprintf("^%s/pods/%s_%s_%s/%s/[0-9]*\\.log", basePath, pod.Namespace, pod.Name, uid, regexp.QuoteMeta(containerName))
	// HACK: try the k8s 1.10 log pattern next, then fall back to our original log pattern
	re2 := fmt.Sprintf("^%s/pods/%s/%s/[0-9]*\\.log", basePath, uid, regexp.QuoteMeta(containerName))
	re3 := fmt.Sprintf("^%s/pods/%s/%s_[0-9]*\\.log", basePath, uid, regexp.QuoteMeta(containerName))
	logrus.WithFields(logrus.Fields{
		"regex1": re1,
		"regex2": re2,
		"regex3": re3,
	}).Debug("Container filter function")

	return func(fileName string) bool {
		// first check if we should exclude this path
		// If exclusionFilter says no, we're going to say no.
		if exclusionFilter != nil && !exclusionFilter(fileName) {
			return false
		}
		// now check our patterns -- if any of them succeeds, we're good
		ok, _ := regexp.MatchString(re1, fileName)
		if ok {
			return ok
		}
		ok, _ = regexp.MatchString(re2, fileName)
		if ok {
			return ok
		}
		ok, _ = regexp.MatchString(re3, fileName)
		return ok
	}
}

func (pt *PodSetTailer) watcherForPod(pod *v1.Pod, containerName string, podWatcher k8sagent.PodWatcher) (*tailer.PathWatcher, error) {
	// at the time we try and start the watcher, the log directory contents may be insufficient
	// to determine the right log pattern. Rather than fail, we just defer this until the watcher is running
	patternFunc := func() (string, error) {
		return determineLogPattern(pod, logsBasePath, pt.legacyLogPaths)
	}

	// only watch logs for containers matching the given name, if
	// one is specified
	filterFunc := determineFilterFunc(pod, logsBasePath, containerName, pt.legacyLogPaths, pt.config.ExcludePaths)

	additionalProcessors := []processors.Processor{
		&processors.KubernetesMetadataProcessor{
			PodGetter: podWatcher,
			UID:       pod.UID,
		},
	}
	// the additionalFields section at the root of the config applies to all pod watchers
	if pt.additionalFieldsGlobal != nil {
		additionalProcessors = append(additionalProcessors, &processors.AdditionalFieldsProcessor{AdditionalFields: pt.additionalFieldsGlobal})
	}
	handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
		pt.config,
		&unwrappers.InferUnwrapper{},
		pt.transmitter,
		additionalProcessors...)
	if err != nil {
		// This shouldn't happen, since we check for configuration errors
		// before actually setting up the watcher
		logrus.WithError(err).Error("Error setting up watcher")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"Name": pod.Name,
		"UID":  pod.UID,
	}).Info("Setting up watcher for pod")

	watcher := tailer.NewPathWatcher(patternFunc, filterFunc, handlerFactory, pt.stateRecorder)
	return watcher, nil
}
