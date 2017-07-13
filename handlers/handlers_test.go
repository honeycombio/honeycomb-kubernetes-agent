package handlers

import (
	"testing"
	"time"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"

	yaml "gopkg.in/yaml.v2"

	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
	"github.com/stretchr/testify/assert"
)

type MockTransmitter struct {
	events []*event.Event
}

func (mt *MockTransmitter) Send(ev *event.Event) {
	mt.events = append(mt.events, ev)
}

func TestInvalidConfigurations(t *testing.T) {
	mt := &MockTransmitter{}

	testcases := []struct {
		config string
		errMsg string
	}{
		{"parser: json", "Missing dataset in configuration"},
		{"dataset: kubernetestest", "No parser specified"},
		{"parser: watparser\ndataset: kubernetestest", "Error setting up parser: Unknown parser type watparser"},
	}

	for _, tc := range testcases {
		cfg := &config.WatcherConfig{}
		err := yaml.Unmarshal([]byte(tc.config), cfg)
		assert.NoError(t, err)

		_, err = NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
		assert.Equal(t, err.Error(), tc.errMsg)
	}
}

func TestDefaultNginxHandling(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "nginx"},
	}
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.DockerJSONLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"log":"192.168.143.128 - - [10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200 612 \"-\" \"curl/7.38.0\" \"-\"\n","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data: map[string]interface{}{
			"bytes_sent":      int64(612),
			"http_user_agent": "curl/7.38.0",
			"remote_addr":     "192.168.143.128",
			"request":         "GET / HTTP/1.1",
			"status":          int64(200),
			"time_local":      "10/Jul/2017:22:10:25 +0000",
		},
		Dataset:   "kubernetestest",
		Path:      "/tmp/testpath",
		Timestamp: time.Date(2017, 7, 10, 22, 10, 25, 569584932, time.UTC),
	}
	assert.Equal(t, mt.events[0], expected)

}

func TestDropField(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "json"},
		Processors: []map[string]map[string]interface{}{
			map[string]map[string]interface{}{
				"drop_field": map[string]interface{}{"field": "todrop"},
			},
		},
	}

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"todrop": "a", "dontdrop": "b"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data:    map[string]interface{}{"dontdrop": "b"},
		Dataset: "kubernetestest",
		Path:    "/tmp/testpath",
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestStaticSampling(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "json"},
		Processors: []map[string]map[string]interface{}{
			map[string]map[string]interface{}{
				"sample": map[string]interface{}{"rate": 10},
			},
		},
	}

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	for i := 0; i < 10000; i++ {
		handler.Handle(`{"field": "a"}`)
	}
	assert.InDelta(t, len(mt.events), 1000, 500)
	for _, ev := range mt.events {
		assert.Equal(t, ev.SampleRate, uint(10))
	}
}

// An (incomplete) test for attaching pod metadata:

type mockPodGetter struct{}

func (mp *mockPodGetter) Get(uid types.UID) (*v1.Pod, bool) {
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "examplePod",
			UID:  uid,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: "containerImage",
				},
				{
					Name:  "sidecar",
					Image: "sidecarImage",
				},
			},
		},
	}, true
}

func TestKubernetesMetadata(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "json"},
	}
	k8sProcessor := &processors.KubernetesMetadataProcessor{
		UID:       types.UID("examplePodUID"),
		PodGetter: &mockPodGetter{},
	}
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt, k8sProcessor)
	assert.NoError(t, err)

	handler := hf.New("/var/log/pods/examplePodUID/container_0.log")
	handler.Handle(`{"field": "a"}`)
	assert.Equal(t, len(mt.events), 1)

	assert.Equal(t, mt.events[0].Data["kubernetes.container.name"], "container")
	assert.Equal(t, mt.events[0].Data["kubernetes.container.image"], "containerImage")
	assert.Equal(t, mt.events[0].Data["kubernetes.pod.UID"], "examplePodUID")
}
