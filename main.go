package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/service"

	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/podtailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/tailer"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/version"
	"github.com/honeycombio/libhoney-go"
	flag "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type OutputSplitter struct{}

// Some systems that process logs expect warn, info, debug, and trace to be on stdout
// and all error output to go to stderr.
func (splitter *OutputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=debug")) || bytes.Contains(p, []byte("level=info")) ||
		bytes.Contains(p, []byte("level=trace")) || bytes.Contains(p, []byte("level=warn")) {
		return os.Stdout.Write(p)
	}
	return os.Stderr.Write(p)
}

type CmdLineOptions struct {
	ConfigPath string `long:"config" description:"Path to configuration file" default:"/etc/honeycomb/config.yaml"`
	Validate   bool   `long:"validate" description:"Validate configuration and exit"`
}

func init() {
	// set the version string to our desired format
	// version.VERSION is the importPath.name ld flag specified by build.sh
	if version.VERSION == "" {
		version.VERSION = "dev"
	}
	// init libhoney user agent properly
	libhoney.UserAgentAddition = fmt.Sprintf("kubernetes/" + version.VERSION)
	logrus.WithFields(logrus.Fields{
		"version": version.VERSION,
	}).Info("Initializing agent")
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		logrus.WithError(err).Error("Error parsing options")
	}
	cfg, err := config.ReadFromFile(flags.ConfigPath)
	if err != nil {
		logrus.WithError(err).Fatal("Error reading configuration")
		os.Exit(1)
	}

	if flags.Validate {
		logrus.Println("Configuration looks good!")
		os.Exit(0)
	}

	if cfg.SplitLogging {
		logrus.SetOutput(&OutputSplitter{})
		logrus.Info("Configured split logging. trace, debug, info, and warn levels will now go to stdout")
	}

	if cfg.Verbosity != "" {
		level, err := logrus.ParseLevel(cfg.Verbosity)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"newLevel": level.String(),
			}).Info("Setting log verbosity")
			logrus.SetLevel(level)
		}
	}

	// Read write key from environment if not specified in config file.
	// k8s secrets injected into the environment are liable to end up with a
	// trailing newline, so trim that.
	// (That's because 'echo "KEY" | base64' encodes KEY plus a trailing
	// newline.)
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("HONEYCOMB_APIKEY"))
		if apiKey == "" {
			// Backwards compatibility check for WriteKey
			apiKey = cfg.WriteKey
			if apiKey == "" {
				apiKey = strings.TrimSpace(os.Getenv("HONEYCOMB_WRITEKEY"))
			}
		}
	}

	err = transmission.InitLibhoney(apiKey, cfg.APIHost, cfg.RetryBufferSize, cfg.RetryBufferExpire)
	if err != nil {
		logrus.WithError(err).Fatal("Error initializing Honeycomb transmission")
	}

	if cfg.Metrics != nil {
		if cfg.Metrics.AdditionalFields == nil && cfg.AdditionalFields != nil {
			cfg.Metrics.AdditionalFields = cfg.AdditionalFields
		}

		err = startMetricsService(cfg.Metrics)
		if err != nil {
			logrus.WithError(err).Fatal("Error while starting metrics service")
		}
	}

	if len(cfg.Watchers) > 0 {
		err = validateWatchers(cfg.Watchers)

		if err != nil {
			// TODO: it'd be really nice to reference the specific configuration
			// block that's problematic when returning an error to the user.
			logrus.WithError(err).Fatal("Error in watcher configuration")
		} else {

			pws, pts := createLogTailers(cfg)
			for _, pw := range pws {
				pw.Start()
				defer pw.Stop()
			}

			for _, pt := range pts {
				pt.Start()
				defer pt.Stop()
			}
		}
	}

	logrus.Info("running")
	waitForSignal()
}

func createLogTailers(config *config.Config) ([]*tailer.PathWatcher, []*podtailer.PodSetTailer) {
	transmitter := &transmission.HoneycombTransmitter{}

	kubeClient, err := newKubeClient()
	if err != nil {
		logrus.WithError(err).Fatal("Error instantiating kube client")
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		logrus.Fatal("No node name set!\n")
	}
	nodeSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)

	stateRecorder, err := tailer.NewStateRecorder("/var/log/honeycomb-agent.state")
	if err != nil {
		logrus.WithError(err).Error("Error initializing state recorder. Agent progress won't be persisted across restarts.")
	}

	pws := make([]*tailer.PathWatcher, 0)
	pts := make([]*podtailer.PodSetTailer, 0)

	for _, watcherConfig := range config.Watchers {
		for _, path := range watcherConfig.FilePaths {
			logrus.WithFields(logrus.Fields{
				"path": path,
			}).Debug("FilePath specified in config")
			handlerFactory, err := handlers.NewLineHandlerFactoryFromConfig(
				watcherConfig,
				&unwrappers.RawLogUnwrapper{},
				transmitter)
			if err != nil {
				// This shouldn't happen, since we check for configuration errors
				// before actually setting up the watcher
				logrus.WithError(err).Error("Error setting up watcher")
				continue
			}
			// Even though we have a static path, NewPathWatcher expects a function,
			// so we build one that just returns the path
			patternFunc := func() (string, error) { return path, nil }
			t := tailer.NewPathWatcher(patternFunc, nil, handlerFactory, stateRecorder)
			pws = append(pws, t)
		}

		if watcherConfig.LabelSelector != nil {
			pt := podtailer.NewPodSetTailer(
				watcherConfig,
				nodeSelector,
				transmitter,
				stateRecorder,
				kubeClient,
				config.LegacyLogPaths,
				config.AdditionalFields,
			)
			pts = append(pts, pt)
		}
	}
	return pws, pts
}

func startMetricsService(config *config.MetricsConfig) error {
	if config.Enabled {

		kubeClient, err := newKubeClient()
		if err != nil {
			logrus.WithError(err).Fatal("Error instantiating kube client")
		}

		if config.Interval == 0 {
			config.Interval = 10 * time.Second
		}

		if config.ClusterName == "" {
			config.ClusterName = "k8s-cluster"
		}

		if config.Dataset == "" {
			config.Dataset = "kubernetes-metrics"
		}

		svc, err := service.NewMetricsService(config, kubeClient)
		if err != nil {
			return err
		}

		if err := svc.Start(); err != nil {
			return err
		}
	}
	return nil
}

func newKubeClient() (*corev1.CoreV1Client, error) {
	// Get client to query API server.
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return corev1.NewForConfig(kubeClientConfig)
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
