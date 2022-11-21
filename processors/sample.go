package processors

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	dynsampler "github.com/honeycombio/dynsampler-go"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
)

type SampleType string

const (
	SampleTypeStatic  SampleType = "static"
	SampleTypeDynamic SampleType = "dynamic"
)

type Sampler struct {
	config     *samplerConfig
	dynsampler dynsampler.Sampler
}

type samplerConfig struct {
	Type            SampleType
	Rate            uint
	Keys            []string
	WindowSize      int
	MinEventsPerSec int
}

func (s *Sampler) Init(options map[string]interface{}) error {
	config := &samplerConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	if config.Type == "" {
		// Default to static if not otherwise specified
		config.Type = SampleTypeStatic
	}
	if config.Type != SampleTypeStatic && config.Type != SampleTypeDynamic {
		return errors.New("sample type must be either 'static' or 'dynamic'")
	}
	if config.WindowSize == 0 {
		// Default to 30 seconds if not otherwise specified
		config.WindowSize = 30
	}
	// MinEventsPerSec is defaulted to 50 by the sampler itself, if the value
	// specified in our config is 0 or missing.

	s.config = config

	if s.config.Type == SampleTypeDynamic {
		s.dynsampler = &dynsampler.AvgSampleWithMin{
			GoalSampleRate:    int(config.Rate),
			ClearFrequencySec: config.WindowSize,
			MinEventsPerSec:   config.MinEventsPerSec,
		}
		if err := s.dynsampler.Start(); err != nil {
			return fmt.Errorf("error starting dynamic sampler: %v", err)
		}
	}
	return nil
}

func (s *Sampler) Process(ev *event.Event) bool {
	var rate uint
	if s.config.Type == SampleTypeStatic {
		rate = s.config.Rate
	} else {
		key := makeDynSampleKey(ev, s.config.Keys)
		rate = uint(s.dynsampler.GetSampleRate(key))
	}
	ev.SampleRate = rate
	return !shouldDrop(rate)

}

func shouldDrop(rate uint) bool {
	return rand.Intn(int(rate)) != 0
}

// From honeytail
func makeDynSampleKey(ev *event.Event, keys []string) string {
	key := make([]string, len(keys))
	for i, field := range keys {
		if val, ok := ev.Data[field]; ok {
			switch val := val.(type) {
			case bool:
				key[i] = strconv.FormatBool(val)
			case int:
				key[i] = strconv.Itoa(val)
			case int64:
				key[i] = strconv.FormatInt(val, 10)
			case float64:
				key[i] = strconv.FormatFloat(val, 'E', -1, 64)
			case string:
				key[i] = val
			default:
				key[i] = "" // skip it
			}
		}
	}
	return strings.Join(key, "_")
}
