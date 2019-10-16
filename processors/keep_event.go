package processors

import (
	"fmt"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type EventKeeper struct {
	config *eventKeeperConfig
	values map[string]struct{}
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

	f.values = make(map[string]struct{}, len(f.config.Values))
	for _, val := range f.config.Values {
		f.values[val] = struct{}{}
	}
	return nil
}

func (f *EventKeeper) Process(ev *event.Event) bool {
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
			"type":  fmt.Sprintf("%T", val),
		}).
			Debug("Not filtering field of non-string type")
	}
	_, exists := f.values[valString]
	return exists
}
