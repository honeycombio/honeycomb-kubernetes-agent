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
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	libhoney "github.com/honeycombio/libhoney-go"
	flag "github.com/jessevdk/go-flags"

	"k8s.io/client-go/kubernetes"
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
		fmt.Println("Setting log level")
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Read write key from environment if not specified in config file
	if config.WriteKey == "" {
		config.WriteKey = os.Getenv("HONEYCOMB_WRITEKEY")
	}

	// k8s secrets are liable to end up with a trailing newline, so trim that.
	err = libhoney.Init(libhoney.Config{
		WriteKey: strings.TrimSpace(config.WriteKey),
	})
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

	go readResponses()

	for _, watcherConfig := range config.Watchers {
		parserFactory, err := parsers.NewParserFactory(watcherConfig.Parser)
		if err != nil {
			fmt.Printf("Error instantiating parser:\n\t%v\n", err)
			os.Exit(1)
		}

		for _, path := range watcherConfig.FilePaths {
			handlerFactory := &handlers.DockerJSONLogHandlerFactory{
				Config:        watcherConfig,
				ParserFactory: parserFactory,
			}

			go tailer.NewPathWatcher(path, nil, handlerFactory).Run()
		}

		if watcherConfig.LabelSelector != nil {
			podWatcher := k8sagent.NewPodWatcher(
				watcherConfig.Namespace,
				*watcherConfig.LabelSelector,
				nodeSelector,
				kubeClient)

			containerName := watcherConfig.ContainerName

			for pod := range podWatcher.Pods() {
				handlerFactory := &handlers.DockerJSONLogHandlerFactory{
					Config:        watcherConfig,
					ParserFactory: parserFactory,
				}

				k := &processors.KubernetesMetadataProcessor{
					PodGetter:     podWatcher,
					ContainerName: containerName,
					UID:           pod.UID}
				handlerFactory.AddProcessor(k)
				pattern := fmt.Sprintf("/var/log/pods/%s/*", pod.UID)
				var filterFunc func(fileName string) bool

				if watcherConfig.ContainerName != "" {
					// only watch logs for containers matching the given name, if
					// one is specified
					re := fmt.Sprintf("^%s_[0-9]*\\.log", regexp.QuoteMeta(containerName))
					filterFunc = func(fileName string) bool {
						ok, _ := regexp.Match(re, []byte(fileName))
						return ok
					}
				}
				pathWatcher := tailer.NewPathWatcher(pattern, filterFunc, handlerFactory)
				go pathWatcher.Run()
			}
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

func readResponses() {
	for resp := range libhoney.Responses() {
		if resp.Err != nil || (resp.StatusCode != 200 && resp.StatusCode != 202) {
			logrus.WithFields(logrus.Fields{
				"error":        resp.Err,
				"status":       resp.StatusCode,
				"responseBody": string(resp.Body),
			}).Error("Failed to send event to Honeycomb")
		}
	}
}
