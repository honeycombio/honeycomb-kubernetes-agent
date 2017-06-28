package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	APIHost   string `yaml:"apiHost"`
	WriteKey  string `yaml:"writekey"`
	Watchers  []*WatcherConfig
	Verbosity string
}

type WatcherConfig struct {
	Parser        string
	Dataset       string
	SampleRate    int `yaml:"sampleRate"`
	Namespace     string
	LabelSelector string `yaml:"labelSelector"`
	ContainerName string `yaml:"containerName"`
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
