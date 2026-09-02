package processors

import (
	"errors"
	"fmt"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var (
	ErrMissingFieldDatasetFromField = errors.New("dataset_from_field requires field to be set")
)

type DatasetResolver struct {
	config *datasetResolverConfig
}

type datasetResolverConfig struct {
	Field string
}

func (f *DatasetResolver) Init(options map[string]interface{}) error {
	config := &datasetResolverConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}

	if config.Field == "" {
		return ErrMissingFieldDatasetFromField
	}
	f.config = config
	return nil
}

func (f *DatasetResolver) Process(ev *event.Event) bool {
	if ev.Data == nil {
		return true
	}
	val, ok := ev.Data[f.config.Field]
	if !ok {
		return true
	}

	valString, ok := val.(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"key":   f.config.Field,
			"value": val,
			"type":  fmt.Sprintf("%T", val)}).
			Debug("Not setting dataset from field of non-string type")
		return true
	}
	if valString != "" {
		ev.Dataset = valString
	}
	return true
}
