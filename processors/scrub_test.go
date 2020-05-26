package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
)

func TestScrubField(t *testing.T) {
	fp := FieldScrubber{}
	fp.Init(map[string]interface{}{"Field": "foo"})

	ev := &event.Event{}
	ev.Data = map[string]interface{}{"foo": "hidden"}

	fp.Process(ev)

	scrubbedVal, ok := ev.Data["foo"]

	assert.True(t, ok)

	assert.Equal(t, scrubbedVal, "e564b4081d7a9ea4b00dada53bdae70c99b87b6fce869f0c3dd4d2bfa1e53e1c")
}
