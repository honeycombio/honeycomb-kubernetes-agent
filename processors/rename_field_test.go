package processors

import (
	"testing"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestRenameField(t *testing.T) {
	// ensures FieldRenamer implements the Processor interface
	var processor Processor
	processor = &FieldRenamer{}

	err := processor.Init(map[string]interface{}{
		"original": "time",
		"new":      "timestamp",
	})
	assert.Equal(t, nil, err, "init should not error")

	e := &event.Event{
		Data: map[string]interface{}{
			"time": "tomorrow",
			"msg":  "leave me be",
		},
	}
	cont := processor.Process(e)
	assert.Equal(t, true, cont, "Process should return true, to signal continued processing")
	assert.Equal(t, nil, e.Data["time"], "time field should be removed")
	assert.Equal(t, "tomorrow", e.Data["timestamp"], "timestamp field should be present")
	assert.Equal(t, "leave me be", e.Data["msg"], "msg field should be unchanged")

	// we should not crash if no data is present
	cont = processor.Process(&event.Event{Data: map[string]interface{}{}})
	assert.Equal(t, true, cont, "Process should return true, to signal continued processing")
}

func TestRenameFieldOverwritesExisting(t *testing.T) {
	// ensures FieldRenamer implements the Processor interface
	var processor Processor
	processor = &FieldRenamer{}

	err := processor.Init(map[string]interface{}{
		"original": "time",
		"new":      "timestamp",
	})
	assert.Equal(t, nil, err, "init should not error")

	e := &event.Event{
		Data: map[string]interface{}{
			"time":      "tomorrow",
			"timestamp": "yesterday",
		},
	}
	cont := processor.Process(e)
	assert.Equal(t, true, cont, "Process should return true, to signal continued processing")
	assert.Equal(t, nil, e.Data["time"], "time field should be removed")
	assert.Equal(t, "tomorrow", e.Data["timestamp"], "timestamp field should have new value")
}

func TestRenameFieldInvalidConfig(t *testing.T) {
	processor := &FieldRenamer{}
	err := processor.Init(map[string]interface{}{
		"original": "time",
		"new":      "time",
	})
	assert.Equal(t, ErrFieldOptionsMatch, err, "matching new and original field names should return an error")
	err = processor.Init(map[string]interface{}{
		"original": "time",
	})
	assert.Equal(t, ErrFieldOptionUnspecified, err, "an error should be returned if new is not specified")
	err = processor.Init(map[string]interface{}{
		"new": "time",
	})
	assert.Equal(t, ErrFieldOptionUnspecified, err, "an error shuould be returned if original is not specified")
}
