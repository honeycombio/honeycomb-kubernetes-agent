// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package service

import (
	"errors"
	"os"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/interval"
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var defaultOmitLabels []metrics.OmitLabel

var defaultMetricGroups = []metrics.MetricGroup{
	//kubelet.ContainerMetricGroup,
	metrics.PodMetricGroup,
	metrics.NodeMetricGroup,
	//kubelet.VolumeMetricGroup,
}

type Service struct {
	config     *config.MetricsConfig
	restClient kubelet.RestClient
	runner     *interval.Runner
	options    Options
	client     *corev1.CoreV1Client
}

type Options struct {
	Dataset               string
	Interval              time.Duration
	ClusterName           string
	OmitLabels            []metrics.OmitLabel
	OmitNameSpaces        []string
	AdditionalFields      map[string]interface{}
	IncludeNodeLabels     bool
	MetricGroupsToCollect map[metrics.MetricGroup]bool
}

func NewMetricsService(cfg *config.MetricsConfig, client *corev1.CoreV1Client) (*Service, error) {

	if cfg.Endpoint == "" {
		// Not all Managed K8s offerings allow the nodename to be DNS reachable, try by IP if available.
		if nodeIp := os.Getenv("NODE_IP"); nodeIp != "" {
			cfg.Endpoint = nodeIp + ":10250"
		} else if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
			cfg.Endpoint = nodeName + ":10250"
		}
	}

	rc, err := createRestClient(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.OmitLabels == nil {
		cfg.OmitLabels = defaultOmitLabels
	}

	// create a map[string]bool from a list of strings
	mg, err := getMapFromSlice(cfg.MetricGroups)
	if err != nil {
		return nil, err
	}

	opt := &Options{
		Dataset:               cfg.Dataset,
		Interval:              cfg.Interval,
		ClusterName:           cfg.ClusterName,
		OmitLabels:            cfg.OmitLabels,
		OmitNameSpaces:        cfg.OmitNameSpaces,
		AdditionalFields:      cfg.AdditionalFields,
		IncludeNodeLabels:     cfg.IncludeNodeLabels,
		MetricGroupsToCollect: mg,
	}

	return &Service{
		config:     cfg,
		restClient: rc,
		options:    *opt,
		client:     client,
	}, nil
}

func createRestClient(cfg *config.MetricsConfig) (kubelet.RestClient, error) {
	logrus.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
	}).Debug("Creating REST Client...")

	clientProvider, err := kubelet.NewClientProvider(cfg.Endpoint)
	if err != nil {
		logrus.Error("Could not create Client Provider: ", err)
		return nil, err
	}
	client, err := clientProvider.BuildClient()
	if err != nil {
		logrus.Error("Could not create Client: ", err)
		return nil, err
	}
	rest := kubelet.NewRestClient(client)
	logrus.Debug("REST Client created")
	return rest, nil
}

// getMapFromSlice returns a set of kubelet.MetricGroup values from
// the provided list. Returns an err if invalid entries are encountered.
func getMapFromSlice(collect []metrics.MetricGroup) (map[metrics.MetricGroup]bool, error) {
	if len(collect) == 0 {
		collect = defaultMetricGroups
	}

	out := make(map[metrics.MetricGroup]bool, len(collect))
	for _, c := range collect {
		if !metrics.ValidMetricGroups[c] {
			return nil, errors.New("invalid entry in metric_groups")
		}
		out[c] = true
	}
	return out, nil
}

func (s *Service) Start() error {

	logrus.WithFields(logrus.Fields{
		"interval":          s.options.Interval,
		"clusterName":       s.options.ClusterName,
		"omitLabels":        s.options.OmitLabels,
		"includeNodeLabels": s.options.IncludeNodeLabels,
		"metricGroups":      s.options.MetricGroupsToCollect,
	}).Info("Creating Metrics Service Runner...")

	// setup primary interval runner for metrics service
	runnable := newRunnable(s.restClient, s.options, s.client)
	s.runner = interval.NewRunner("k8s-stats", s.options.Interval, runnable)
	logrus.Debug("Metrics Service Runner created")

	// start metrics service in go routine
	go func() {
		if err := s.runner.Start(); err != nil {
			logrus.WithError(err).Error("Failed to start Metrics Service")
		}
	}()
	return nil
}

func (s *Service) Shutdown() {
	s.runner.Stop()
}
