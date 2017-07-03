package processors

import (
	"errors"
	"fmt"
)

type Processor interface {
	Process(data map[string]interface{})
	Init(options map[string]interface{}) error
}

func NewProcessor(name string, options map[string]interface{}) (Processor, error) {
	switch name {
	case "request_shape":
		p := &RequestShaper{}
		err := p.Init(options)
		return p, err
	}
	return nil, fmt.Errorf("Unknown processor type %s", name)
}

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
