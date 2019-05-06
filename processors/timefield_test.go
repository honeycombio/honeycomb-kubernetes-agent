package processors

import (
	"testing"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

// Ensure that the timestamp processor can handle timestamp fields of type time.Time
func TestTimefieldProcessReturnsTime(t *testing.T) {
	// Format should be irrelevant here, we just return the field if it's there
	tp := TimeFieldExtractor{config: &timeFieldConfig{Field: "special_time_field", Format: time.RFC3339}}

	mockTime := time.Now()
	ev := &event.Event{}
	assert.True(t, ev.Timestamp.IsZero())
	ev.Data = map[string]interface{}{"foo": "some_value", "special_time_field": mockTime}
	tp.Process(ev)
	assert.Equal(t, mockTime, ev.Timestamp)
}
