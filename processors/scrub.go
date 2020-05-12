package processors

import (
	"crypto/sha256"
	"fmt"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
)

type FieldScrubber struct {
	config *fieldScrubberConfig
}

type fieldScrubberConfig struct {
	Field string
}

func (f *FieldScrubber) Init(options map[string]interface{}) error {
	config := &fieldScrubberConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	f.config = config
	return nil
}

func (f *FieldScrubber) Process(ev *event.Event) bool {
	if val, ok := ev.Data[f.config.Field]; ok {
		// generate a sha256 hash and use the base16 for the content
		newVal := sha256.Sum256([]byte(fmt.Sprintf("%v", val)))
		ev.Data[f.config.Field] = fmt.Sprintf("%x", newVal)
	}
	return true
}
