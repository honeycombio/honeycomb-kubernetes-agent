package processors

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
)

type AdditionalFieldsProcessor struct {
	AdditionalFields map[string]interface{}
}

func (a *AdditionalFieldsProcessor) Init(options map[string]interface{}) error {
	// options expects a map of string->values
	if options != nil {
		a.AdditionalFields = options
	} else {
		a.AdditionalFields = make(map[string]interface{})
	}
	return nil
}

func (a *AdditionalFieldsProcessor) Process(ev *event.Event) bool {
	for k, v := range a.AdditionalFields {
		ev.Data[k] = v
	}

	return true
}
