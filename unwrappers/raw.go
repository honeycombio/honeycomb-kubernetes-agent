package unwrappers

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
)

type RawLogUnwrapper struct{}

func (u *RawLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (*event.Event, error) {

	data, err := parser.Parse(rawLine)
	if err != nil {
		return nil, err
	}
	return &event.Event{Data: data}, nil
}
