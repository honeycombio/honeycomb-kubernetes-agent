package processors

import (
	"testing"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestSamplerCanBuildsKeysWithInts(t *testing.T) {
	key := makeDynSampleKey(&event.Event{
		Data: map[string]interface{}{
			"status": "200",
		},
	}, []string { "status"})
	assert.Equal(t, "200", key)
}
