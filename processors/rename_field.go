package processors

import (
	"errors"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrFieldOptionUnspecified = errors.New("rename_field processor requires both 'new' and 'original' field names to be set")
	ErrFieldOptionsMatch      = errors.New("rename_field processor does not support matching 'new' and 'original' field names")
)

type FieldRenamer struct {
	config *fieldRenamerConfig
}

type fieldRenamerConfig struct {
	Original string
	New      string
}

func (f *FieldRenamer) Init(options map[string]interface{}) error {
	config := &fieldRenamerConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}

	if config.New == "" || config.Original == "" {
		return ErrFieldOptionUnspecified
	}
	if config.New == config.Original {
		return ErrFieldOptionsMatch
	}
	f.config = config
	return nil
}

func (f *FieldRenamer) Process(ev *event.Event) bool {
	if ev.Data != nil {
		if field_data, found := ev.Data[f.config.Original]; found {
			ev.Data[f.config.New] = field_data
			delete(ev.Data, f.config.Original)
		}
	}
	return true
}
