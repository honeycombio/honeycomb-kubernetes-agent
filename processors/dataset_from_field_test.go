package processors

import (
	"testing"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestDatasetFromField(t *testing.T) {
	var processor Processor
	processor = &DatasetResolver{}

	err := processor.Init(map[string]interface{}{"field": "service.name"})
	assert.Equal(t, nil, err, "init should not error")

	e := &event.Event{
		Dataset: "fallback",
		Data:    map[string]interface{}{"service.name": "fintech-connector"},
	}
	cont := processor.Process(e)
	assert.Equal(t, true, cont, "Process should return true, to signal continued processing")
	assert.Equal(t, "fintech-connector", e.Dataset, "dataset should come from the field")

	cont = processor.Process(&event.Event{Data: map[string]interface{}{}})
	assert.Equal(t, true, cont, "Process should return true, to signal continued processing")
}

func TestDatasetFromFieldLeavesDatasetAlone(t *testing.T) {
	processor := &DatasetResolver{}
	err := processor.Init(map[string]interface{}{"field": "service.name"})
	assert.Equal(t, nil, err, "init should not error")

	testCases := []struct {
		name string
		data map[string]interface{}
	}{
		{"field absent", map[string]interface{}{"msg": "hi"}},
		{"field empty", map[string]interface{}{"service.name": ""}},
		{"field not a string", map[string]interface{}{"service.name": 42}},
		{"no data", nil},
	}

	for _, tc := range testCases {
		e := &event.Event{Dataset: "fallback", Data: tc.data}
		cont := processor.Process(e)
		assert.Equal(t, true, cont, tc.name+": Process should return true")
		assert.Equal(t, "fallback", e.Dataset, tc.name+": dataset should be unchanged")
	}
}

func TestDatasetFromFieldRequiresField(t *testing.T) {
	processor := &DatasetResolver{}
	err := processor.Init(map[string]interface{}{})
	assert.Equal(t, ErrMissingFieldDatasetFromField, err, "init should require field")
}
