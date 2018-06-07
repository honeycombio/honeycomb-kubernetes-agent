package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/podtailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	flag "github.com/jessevdk/go-flags"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CmdLineOptions struct {
	ConfigPath string `long:"config" description:"Path to configuration file" default:"/etc/honeycomb/config.yaml"`
	Validate   bool   `long:"validate" description:"Validate configuration and exit"`
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		fmt.Printf("Error parsing options:\n%v\n", err)
	}
	config, err := config.ReadFromFile(flags.ConfigPath)
	if err != nil {
		fmt.Printf("Error reading configuration:\n%v\n", err)
		os.Exit(1)
	}

	if len(config.Watchers) == 0 {
		fmt.Printf("No watchers defined in the configuration!")
		os.Exit(1)
	}

	err = validateWatchers(config.Watchers)
	if err != nil {
		// TODO: it'd be really nice to reference the specific configuration
		// block that's problematic when returning an error to the user.
		fmt.Printf("Error in watcher configuration:\n%v\n", err)
		os.Exit(1)
	}

	if config.Verbosity == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if flags.Validate {
		fmt.Println("Configuration looks good!")
		os.Exit(0)
	}

	// Read write key from environment if not specified in config file.
	// k8s secrets injected into the environment are liable to end up with a
	// trailing newline, so trim that.
	// (That's because 'echo "KEY" | base64' encodes KEY plus a trailing
	// newline.)
	if config.WriteKey == "" {
		config.WriteKey = strings.TrimSpace(os.Getenv("HONEYCOMB_WRITEKEY"))
	}

	err = transmission.InitLibhoney(config.WriteKey, config.APIHost)
	if err != nil {
		fmt.Printf("Error initializing Honeycomb transmission:\n%v\n", err)
		os.Exit(1)
	}
	transmitter := &transmission.HoneycombTransmitter{}

	kubeClient, err := newKubeClient()
	if err != nil {
		fmt.Printf("Error instantiating kube client:\n%v\n", err)
		os.Exit(1)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		fmt.Printf("No node name set!\n")
		os.Exit(1)
	}
	nodeSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)

	stateRecorder, err := tailer.NewStateRecorder("/var/log/honeycomb-agent.state")
	if err != nil {
		logrus.WithError(err).Error("Error initializing state recorder. Agent progress won't be persisted across restarts.")
	}

	for _, watcherConfig := range config.Watchers {
		for _, path := range watcherConfig.FilePaths {
			handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
				watcherConfig,
				&unwrappers.RawLogUnwrapper{},
				transmitter)
			if err != nil {
				// This shouldn't happen, since we check for configuration errors
				// before actually setting up the watcher
				logrus.WithError(err).Error("Error setting up watcher")
			}
			t := tailer.NewPathWatcher(path, nil, handlerFactory, stateRecorder)
			t.Start()
			defer t.Stop()
		}

		if watcherConfig.LabelSelector != nil {
			pt := podtailer.NewPodSetTailer(
				watcherConfig,
				nodeSelector,
				transmitter,
				stateRecorder,
				kubeClient,
			)
			pt.Start()
			defer pt.Stop()
		}
	}

	fmt.Println("running")
	waitForSignal()
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
			return options, fmt.Errorf("Unexpected extra arguments: %s\n", strings.Join(extraArgs, " "))
		}
	}
	return options, nil
}

func validateWatchers(configs []*config.WatcherConfig) error {
	for _, watcherConfig := range configs {
		_, err := handlers.NewLineHandlerFactoryFromConfig(
			watcherConfig,
			&unwrappers.RawLogUnwrapper{},
			&transmission.NullTransmitter{},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
}
