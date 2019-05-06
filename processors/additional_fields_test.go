package processors

import (
	"testing"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestAdditionalFieldsProcess(t *testing.T) {
	fields := map[string]interface{}{"foo": 1, "bar": "two"}

	p := &AdditionalFieldsProcessor{}
	p.Init(fields)

	ev := &event.Event{}
	ev.Data = make(map[string]interface{})
	p.Process(ev)
	assert.Equal(t, ev.Data["foo"], 1)
	assert.Equal(t, ev.Data["bar"], "two")
}

func TestAdditionalFieldsProcessNilMap(t *testing.T) {
	p := &AdditionalFieldsProcessor{}
	p.Init(nil)

	ev := &event.Event{}
	ev.Data = make(map[string]interface{})
	// shouldn't crash
	p.Process(ev)
}
