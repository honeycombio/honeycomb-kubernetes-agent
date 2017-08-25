// Package processors contains support for mutating event data after it's been
// parsed out of an event line.
package processors

import (
	"errors"
	"fmt"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
)

// Processor is the interface that processors implement. The Init() method is
// called to initialize the processor. Process() mutates event data in-place.
// Processors should return `true` if processing and sending should continue,
// and `false` if not. This is primarily relevant for sampling, and other
// processors should always return true.
type Processor interface {
	Process(*event.Event) bool
	Init(options map[string]interface{}) error
}

// NewProcessorFromConfig takes a configuration map that's been unmarshalled
// out of YAML, and tries to instantiate a corresponding parser.
// The syntax for processor configuration is:
// processors:
// - request_shape:
//     field: request
//     prefix: shaped
// or equivalently:
// {"processors": [{"request_shape": {"field": "request", "prefix": "shaped"}}]}
// So NewProcessorFromConfig expects to get a map with one key (the name of the
// processor).
func NewProcessorFromConfig(config map[string]map[string]interface{}) (Processor, error) {
	if len(config) != 1 {
		// TODO: better error
		return nil, fmt.Errorf("Invalid processor configuration")
	}
	for name, options := range config {
		return NewProcessor(name, options)
	}

	return nil, errors.New("No processor found")
}

func NewProcessor(name string, options map[string]interface{}) (Processor, error) {
	var p Processor
	switch name {
	case "request_shape":
		p = &RequestShaper{}
	case "drop_field":
		p = &FieldDropper{}
	case "sample":
		p = &Sampler{}
	case "timefield":
		p = &TimeFieldExtractor{}
	default:
		return nil, fmt.Errorf("Unknown processor type %s", name)
	}
	err := p.Init(options)
	return p, err
}
