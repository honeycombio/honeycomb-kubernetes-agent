package handlers

import (
	"testing"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	"github.com/stretchr/testify/assert"
)

type MockTransmitter struct {
	events []*event.Event
}

func (mt *MockTransmitter) Send(ev *event.Event) {
	mt.events = append(mt.events, ev)
}

func TestDefaultNginxHandling(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "nginx"},
	}
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.DockerJSONLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("test")
	handler.Handle(`{"log":"192.168.143.128 - - [10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200 612 \"-\" \"curl/7.38.0\" \"-\"\n","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Timestamp: time.Date(2017, 7, 10, 22, 10, 25, 569584932, time.UTC),
		Dataset:   "kubernetestest",
		Data: map[string]interface{}{
			"bytes_sent":      int64(612),
			"http_user_agent": "curl/7.38.0",
			"remote_addr":     "192.168.143.128",
			"request":         "GET / HTTP/1.1",
			"status":          int64(200),
			"time_local":      "10/Jul/2017:22:10:25 +0000",
		},
	}
	assert.Equal(t, mt.events[0], expected)

}
