package processors

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
)

type FieldDropper struct {
	config *fieldDropperConfig
}

type fieldDropperConfig struct {
	field string
}

func (f *FieldDropper) Init(options map[string]interface{}) error {
	config := &fieldDropperConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	f.config = config
	return nil
}

func (f *FieldDropper) Process(ev *event.Event) {
	delete(ev.Data, f.config.field)
}
