package processors

import (
	"math/rand"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
)

type Sampler struct {
	config *samplerConfig
}

type samplerConfig struct {
	Type string
	Rate uint
}

func (s *Sampler) Init(options map[string]interface{}) error {
	config := &samplerConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

func (s *Sampler) Process(ev *event.Event) bool {
	ev.SampleRate = s.config.Rate
	return !shouldDrop(s.config.Rate)
}

func shouldDrop(rate uint) bool {
	return rand.Intn(int(rate)) != 0
}
