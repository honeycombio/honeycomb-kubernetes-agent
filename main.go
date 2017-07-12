package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/k8sagent"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	flag "github.com/jessevdk/go-flags"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type CmdLineOptions struct {
	ConfigPath string `long:"config" description:"Path to configuration file" default:"/etc/honeycomb/config.yaml"`
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		fmt.Printf("Error parsing options:\n\t%v\n", err)
	}
	config, err := config.ReadFromFile(flags.ConfigPath)
	if err != nil {
		fmt.Printf("Error reading configuration:\n\t%v\n", err)
		os.Exit(1)
	}

	if config.Verbosity == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Read write key from environment if not specified in config file
	if config.WriteKey == "" {
		config.WriteKey = os.Getenv("HONEYCOMB_WRITEKEY")
	}

	// k8s secrets are liable to end up with a trailing newline, so trim that.
	err = transmission.InitLibhoney(strings.TrimSpace(config.WriteKey), config.APIHost)

	transmitter := &transmission.HoneycombTransmitter{}

	if err != nil {
		fmt.Printf("Error initializing Honeycomb transmission:\n\t%v\n", err)
		os.Exit(1)
	}

	if len(config.Watchers) == 0 {
		fmt.Printf("No watchers defined in the configuration!")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error fetching pod list:\n\t%v\n", err)
		os.Exit(1)
	}

	kubeClient, err := newKubeClient()
	if err != nil {
		fmt.Printf("Error instantiating kube client:\n\t%v\n", err)
		os.Exit(1)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		fmt.Printf("No node name set!\n")
		os.Exit(1)
	}
	nodeSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)

	for _, watcherConfig := range config.Watchers {
		for _, path := range watcherConfig.FilePaths {
			handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
				watcherConfig,
				&unwrappers.RawLogUnwrapper{},
				transmitter)
			if err != nil {
				fmt.Printf("Error setting up watcher for path %s:\n\t%v\n",
					path, err)
				os.Exit(1)

			}
			go tailer.NewPathWatcher(path, nil, handlerFactory).Run()
		}

		if watcherConfig.LabelSelector != nil {
			// Check for errors setting up the handler straightaway --
			// we don't need to have actual pods to tail to do this.
			_, err := handlers.NewLineHandlerFactoryFromConfig(
				watcherConfig,
				&unwrappers.DockerJSONLogUnwrapper{},
				transmitter)
			if err != nil {
				fmt.Printf("Error setting up watcher for LabelSelector %s:\n",
					*watcherConfig.LabelSelector)
				fmt.Printf("\t%v\n", err)
				os.Exit(1)
			}

			go watchPods(watcherConfig, nodeSelector, transmitter, kubeClient)
		}
	}

	fmt.Println("running")
	// Hang out forever
	select {}
}

func newKubeClient() (*kubernetes.Clientset, error) {
	// Get clientset to query API server.
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeClientConfig)
}

func parseFlags() (CmdLineOptions, error) {
	var options CmdLineOptions
	flagParser := flag.NewParser(&options, flag.PrintErrors)
	if extraArgs, err := flagParser.Parse(); err != nil || len(extraArgs) != 0 {
		if err != nil {
			return options, err
		} else {
			return options, fmt.Errorf("\tUnexpected extra arguments: %s\n", strings.Join(extraArgs, " "))
		}
	}
	return options, nil
}

func watchPods(
	watcherConfig *config.WatcherConfig,
	nodeSelector string,
	transmitter transmission.Transmitter,
	kubeClient *kubernetes.Clientset,
) {

	podWatcher := k8sagent.NewPodWatcher(
		watcherConfig.Namespace,
		*watcherConfig.LabelSelector,
		nodeSelector,
		kubeClient)

	for pod := range podWatcher.Pods() {
		k8sMetadataProcessor := &processors.KubernetesMetadataProcessor{
			PodGetter:     podWatcher,
			ContainerName: watcherConfig.ContainerName,
			UID:           pod.UID}
		handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
			watcherConfig,
			&unwrappers.DockerJSONLogUnwrapper{},
			transmitter,
			k8sMetadataProcessor)
		if err != nil {
			// This shouldn't happen, since we check for configuration errors
			// before actually setting up the watcher
			logrus.WithError(err).Error("Error setting up watcher")
			continue
		}
		go watchFilesForPod(pod, watcherConfig.ContainerName, handlerFactory)
	}
}

func watchFilesForPod(pod *v1.Pod, containerName string, handlerFactory handlers.LineHandlerFactory) {
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
	pathWatcher := tailer.NewPathWatcher(pattern, filterFunc, handlerFactory)
	pathWatcher.Run()
}
