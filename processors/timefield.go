package processors

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/mitchellh/mapstructure"
)

type timeFieldConfig struct {
	Field  string
	Format string
}

type TimeFieldExtractor struct {
	config *timeFieldConfig
}

func (t *TimeFieldExtractor) Init(options map[string]interface{}) error {
	config := &timeFieldConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	t.config = config
	return nil
}

func (t *TimeFieldExtractor) Process(ev *event.Event) bool {
	ev.Timestamp = httime.GetTimestamp(
		ev.Data,
		t.config.Field,
		t.config.Format)
	return true
}
