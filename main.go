package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
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

	err = libhoney.Init(libhoney.Config{
		WriteKey: config.WriteKey,
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

	for _, watcherConfig := range config.Watchers {
		podWatcher := k8sagent.NewPodWatcher(
			watcherConfig.Namespace,
			watcherConfig.LabelSelector,
			nodeSelector,
			kubeClient)

		containerName := watcherConfig.ContainerName

		for pod := range podWatcher.Pods() {
			k := &processors.KubernetesMetadataProcessor{
				PodGetter:     podWatcher,
				ContainerName: containerName,
				UID:           pod.UID}
			handlerFactory := &handlerFactory{
				config: watcherConfig,
			}
			handlerFactory.AddProcessor(k)
			pattern := fmt.Sprintf("/var/log/pods/%s/*", pod.UID)
			var filterFunc func(fileName string) bool

			if watcherConfig.ContainerName != "" {
				// only watch logs for containers matching the given name, if
				// specified
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

	fmt.Println("running")
	// Hang out forever
	select {}
}

type handlerFactory struct {
	config     *config.WatcherConfig
	processors []Processor
}

func (h *handlerFactory) New(path string) tailer.LineHandler {
	handler := &JSONLogHandler{
		config: h.config,
		parser: &parsers.NoOpParser{},
	}
	for _, p := range h.processors {
		handler.AddProcessor(p)
	}
	handler.Init()
	return handler
}

func (h *handlerFactory) AddProcessor(p Processor) {
	h.processors = append(h.processors, p)
}

func newKubeClient() (*kubernetes.Clientset, error) {
	// Get clientset to query API server.
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeClientConfig)
}

type JSONLogHandler struct {
	config         *config.WatcherConfig
	parser         parsers.Parser
	postprocessors []Processor
	builder        *libhoney.Builder
}

func (h *JSONLogHandler) Init() {
	h.builder = libhoney.NewBuilder()
	h.builder.Dataset = h.config.Dataset
	h.parser.Init()
	// TODO handle postprocessors
}

type jsonLogLine struct {
	Log    string
	Stream string
	Time   string
}

func (h *JSONLogHandler) Handle(rawLine string) {
	// multiline parsing should be done with stateful parsers for now
	line := &jsonLogLine{}
	err := json.Unmarshal([]byte(rawLine), line)
	if err != nil {
		logrus.WithError(err).Info("Error parsing JSON line")
		return
	}

	parsed, err := h.parser.Parse(line.Log)
	if err != nil {
		logrus.WithError(err).Debug("Failed to parse line")
		return
	}
	for _, p := range h.postprocessors {
		p.Process(parsed)
	}
	logrus.WithField("parsed", parsed).Debug("Sending line")
	h.builder.SendNow(parsed)
}

func (h *JSONLogHandler) AddProcessor(p Processor) {
	h.postprocessors = append(h.postprocessors, p)
}

type Processor interface {
	Process(data map[string]interface{})
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
