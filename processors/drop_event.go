package processors

import (
	"errors"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var (
	ErrFilterOptionUnspecified = errors.New("drop_event processor requires a 'Field' to be set")
)

type EventDropper struct {
	config *eventDropperConfig
	values map[string]bool
}

type eventDropperConfig struct {
	Field  string
	Values []string
}

func (f *EventDropper) Init(options map[string]interface{}) error {
	config := &eventDropperConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}

	if config.Field == "" {
		return ErrFilterOptionUnspecified
	}
	f.config = config

	values := make(map[string]bool)
	for _, val := range f.config.Values {
		values[val] = true
	}
	f.values = values
	return nil
}

func (f *EventDropper) Process(ev *event.Event) bool {
	if ev.Data != nil {
		val, ok := ev.Data[f.config.Field]
		if !ok {
			return true
		}
		valString, ok := val.(string)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"key":   f.config.Field,
				"value": val}).
				Debug("Not filtering field of non-string type")
		}
		_, exists := f.values[valString]
		if exists {
			return false
		}
	}
	return true
}
