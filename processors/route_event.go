package processors

import (
	"errors"
	"fmt"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type EventRouter struct {
	config *eventRouterConfig
	routes map[string]string
}

type eventRouterConfig struct {
	Field  string
	Routes []eventRoute
}

type eventRoute struct {
	Dataset string
	Value   string
}

var (
	ErrDuplicateEventRoute    = errors.New("route_event requires all values to be unique")
	ErrMissingFieldEventRoute = errors.New("route_event requires field to be set")
)

func (f *EventRouter) Init(options map[string]interface{}) error {
	config := &eventRouterConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}

	if config.Field == "" {
		return ErrMissingFieldEventRoute
	}
	f.routes = make(map[string]string, len(config.Routes))

	for _, route := range config.Routes {
		// Same route pointing to multiple datasets?
		if _, ok := f.routes[route.Value]; ok {
			return ErrDuplicateEventRoute
		}
		f.routes[route.Value] = route.Dataset
	}
	f.config = config
	return nil
}

func (f *EventRouter) Process(ev *event.Event) bool {
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
			Debug("Not routing field of non-string type")
		return true
	}
	if dataset, ok := f.routes[valString]; ok {
		ev.Dataset = dataset
	}
	return true
}
