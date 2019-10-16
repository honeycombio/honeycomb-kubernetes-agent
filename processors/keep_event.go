package processors

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type EventKeeper struct {
	config *eventKeeperConfig
	values map[string]bool
}

type eventKeeperConfig struct {
	Field  string
	Values []string
}

func (f *EventKeeper) Init(options map[string]interface{}) error {
	config := &eventKeeperConfig{}
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

func (f *EventKeeper) Process(ev *event.Event) bool {
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
			return true
		}
		return false
	}
	return true
}
