// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package service

import (
	"errors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/interval"
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
	"os"
	"time"
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
	builder    *libhoney.Builder
	logger     *logrus.Logger
}

type Options struct {
	Interval              time.Duration
	ClusterName           string
	OmitLabels            []metrics.OmitLabel
	MetricGroupsToCollect map[metrics.MetricGroup]bool
}

func NewMetricsService(cfg *config.MetricsConfig, builder *libhoney.Builder, logger *logrus.Logger) (*Service, error) {

	if cfg.Endpoint == "" {
		if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
			cfg.Endpoint = nodeName + ":10250"
		}
	}

	rc, err := createRestClient(cfg, logger)
	if err != nil {
		return nil, err
	}

	if cfg.OmitLabels == nil {
		cfg.OmitLabels = defaultOmitLabels
	}

	mg, err := getMapFromSlice(cfg.MetricGroups)
	if err != nil {
		return nil, err
	}

	opt := &Options{
		Interval:              cfg.Interval,
		ClusterName:           cfg.ClusterName,
		OmitLabels:            cfg.OmitLabels,
		MetricGroupsToCollect: mg,
	}

	return &Service{
		config:     cfg,
		restClient: rc,
		options:    *opt,
		builder:    builder,
		logger:     logger,
	}, nil
}

func createRestClient(cfg *config.MetricsConfig, logger *logrus.Logger) (kubelet.RestClient, error) {
	logger.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
	}).Debug("Creating REST Client...")

	clientProvider, err := kubelet.NewClientProvider(cfg.Endpoint)
	if err != nil {
		logger.Error("Could not create Client Provider: ", err)
		return nil, err
	}
	client, err := clientProvider.BuildClient()
	if err != nil {
		logger.Error("Could not create Client: ", err)
		return nil, err
	}
	rest := kubelet.NewRestClient(client)
	logger.Debug("REST Client created")
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

	s.logger.WithFields(logrus.Fields{
		"interval":     s.options.Interval,
		"clusterName":  s.options.ClusterName,
		"omitLabels":   s.options.OmitLabels,
		"metricGroups": s.options.MetricGroupsToCollect,
	}).Info("Creating Metrics Service Runner...")
	runnable := newRunnable(s.restClient, s.builder, s.options, s.logger)
	s.runner = interval.NewRunner("k8s-stats", s.options.Interval, runnable)
	s.logger.Debug("Metrics Service Runner created")

	go func() {
		if err := s.runner.Start(); err != nil {
			s.logger.WithError(err).Error("Failed to start Metrics Service")
		}
	}()
	return nil
}

func (s *Service) Shutdown() {
	s.runner.Stop()
}
