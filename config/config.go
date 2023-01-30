package config

import (
	"io/ioutil"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics"

	"github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	APIKey  string `yaml:"apiKey"`
	APIHost string `yaml:"apiHost"`
	// Deprecated: use APIKey instead.
	WriteKey          string        `yaml:"writekey"`
	RetryBufferSize   int           `yaml:"retryBufferSize"`
	RetryBufferExpire time.Duration `yaml:"retryBufferExpire"`
	Watchers          []*WatcherConfig
	Verbosity         string
	LegacyLogPaths    bool                   `yaml:"legacyLogPaths"`
	SplitLogging      bool                   `yaml:"splitLogging"`
	AdditionalFields  map[string]interface{} `yaml:"additionalFields"`
	Metrics           *MetricsConfig
}

type WatcherConfig struct {
	Parser    *ParserConfig
	Dataset   string
	Namespace string
	// Distinguish between nil and empty string in the LabelSelector
	// nil means watch no pods, empty string means watch all of them
	// Maybe we need a better API? But k8s is pretty insistent that empty
	// string means "select all pods".
	LabelSelector *string  `yaml:"labelSelector"`
	FilePaths     []string `yaml:"paths" validate:"excluded_with=LabelSelector"`
	ExcludePaths  []string `yaml:"exclude"`
	ContainerName string   `yaml:"containerName"`
	Processors    []map[string]map[string]interface{}
}

type ParserConfig struct {
	Name    string
	Options map[string]interface{}
}

type MetricsConfig struct {
	Enabled     bool
	Dataset     string
	Endpoint    string
	Interval    time.Duration
	ClusterName string `yaml:"clusterName"`
	// Labels to omit from becoming fields in Honeycomb
	// By default `controller-revision-hash` is omitted
	OmitLabels []metrics.OmitLabel `yaml:"omitLabels"`
	// Include node metadata such as 'node.kubernetes.io/instance-type' or
	// 'topology.kubernetes.io/region'
	IncludeNodeLabels bool `yaml:"includeNodeLabels"`
	// MetricGroupsToCollect provides a list of metrics groups to collect metrics from.
	// "container", "pod", "node" and "volume" are the only valid groups.
	MetricGroups     []metrics.MetricGroup  `yaml:"metricGroups"`
	AdditionalFields map[string]interface{} `yaml:"additionalFields"`
}

func (p *ParserConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var name string
	if err := unmarshal(&name); err == nil {
		p.Name = name
		return nil
	}
	aux := &struct {
		Name    string
		Options map[string]interface{}
	}{}
	err := unmarshal(&aux)
	if err == nil {
		p.Name = aux.Name
		p.Options = aux.Options
		return nil
	}
	return err
}

func ReadFromFile(filePath string) (*Config, error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err = yaml.Unmarshal(contents, config); err != nil {
		return nil, err
	}

	v := validator.New()
	for _, watcher := range config.Watchers {
		err = v.Struct(watcher)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
