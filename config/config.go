package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	APIHost        string `yaml:"apiHost"`
	WriteKey       string `yaml:"writekey"`
	Watchers       []*WatcherConfig
	Verbosity      string
	LegacyLogPaths bool   `yaml:"legacyLogPaths"`
	SplitLogging   bool   `yaml:"splitLogging"`
	ErrorDataset   string `yaml:"errorDataset"`
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
	FilePaths     []string `yaml:"paths"`
	ContainerName string   `yaml:"containerName"`
	Processors    []map[string]map[string]interface{}
	ErrorDataset  string `yaml:"errorDataset"`
}

type ParserConfig struct {
	Name    string
	Options map[string]interface{}
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
	return config, nil
}
