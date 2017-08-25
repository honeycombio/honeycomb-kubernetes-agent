package handlers

// It makes sense for most unit tests to live here, even ones that specifically
// exercise particular processors and such, since they use a common test setup:
// instantiate a LineHandler from a given configuration, pass it some example
// lines, and test that the resultant events match expectations.

import (
	"fmt"
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

func watcherConfigFromYAML(configSnippet string) (*config.WatcherConfig, error) {
	cfg := &config.WatcherConfig{}
	err := yaml.Unmarshal([]byte(configSnippet), cfg)
	return cfg, err
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
		cfg, err := watcherConfigFromYAML(tc.config)
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

func TestGlogParsing(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "glog"},
	}
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.DockerJSONLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"log": "W0720 00:15:01.592300       5 controller.go:386] Resetting endpoints for master service", "stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data: map[string]interface{}{
			"level":          "warning",
			"filename":       "controller.go",
			"lineno":         "386",
			"message":        "Resetting endpoints for master service",
			"threadid":       "5",
			"glog_timestamp": time.Date(time.Now().Year(), 07, 20, 0, 15, 01, 592300000, time.UTC),
		},
		Dataset:   "kubernetestest",
		Path:      "/tmp/testpath",
		Timestamp: time.Date(2017, 7, 10, 22, 10, 25, 569584932, time.UTC),
	}
	assert.Equal(t, mt.events[0], expected)

}

func TestRedisParsing(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "redis"},
	}
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.DockerJSONLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"log": "44:C 09 Aug 23:12:19.127 * RDB: 0 MB of memory used by copy-on-write", "stream":"stdout","time":"2017-07-02T22:10:25.569534932Z"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data: map[string]interface{}{
			"level":           "notice",
			"role":            "child",
			"pid":             "44",
			"message":         "RDB: 0 MB of memory used by copy-on-write",
			"redis_timestamp": time.Date(time.Now().Year(), 8, 9, 23, 12, 19, 127000000, time.UTC),
		},
		Dataset:   "kubernetestest",
		Path:      "/tmp/testpath",
		Timestamp: time.Date(2017, 7, 2, 22, 10, 25, 569534932, time.UTC),
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestKeyvalParsing(t *testing.T) {
	mt := &MockTransmitter{}
	cfg, err := watcherConfigFromYAML(`
---
dataset: kubernetestest
parser:
  name: keyval
  options:
    prefixRegex: "(?P<timestamp>[0-9\\.:TZ\\-]+) AUDIT: "
processors:
  - timefield:
      field: timestamp
`)
	assert.NoError(t, err)
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`2017-08-25T04:40:49.965969122Z AUDIT: id="8e2ad929-8da1-4ce5-8776-751421157483" ip="172.20.67.135" method="GET" user="system:serviceaccount:tectonic-system:prometheus-operator" groups="\"system:serviceaccounts\",\"system:serviceaccounts:tectonic-system\",\"system:authenticated\"" as="<self>" asgroups="<lookup>" namespace="tectonic-system" uri="/api/v1/namespaces/tectonic-system/secrets/prometheus-k8s"`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data: map[string]interface{}{
			"id":        "8e2ad929-8da1-4ce5-8776-751421157483",
			"ip":        "172.20.67.135",
			"method":    "GET",
			"user":      "system:serviceaccount:tectonic-system:prometheus-operator",
			"groups":    "\"system:serviceaccounts\",\"system:serviceaccounts:tectonic-system\",\"system:authenticated\"",
			"as":        "<self>",
			"asgroups":  "<lookup>",
			"namespace": "tectonic-system",
			"uri":       "/api/v1/namespaces/tectonic-system/secrets/prometheus-k8s",
		},
		Dataset:   "kubernetestest",
		Path:      "/tmp/testpath",
		Timestamp: time.Date(2017, 8, 25, 4, 40, 49, 965969122, time.UTC),
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
	cfg, err := watcherConfigFromYAML(`
dataset: kubernetestest
parser: json
processors:
- sample:
    type: static
    rate: 10`)
	assert.NoError(t, err)

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

func TestDynamicSampling(t *testing.T) {
	mt := &MockTransmitter{}
	cfg, err := watcherConfigFromYAML(`
dataset: kubernetestest
parser: json
processors:
- sample:
    type: dynamic
    windowSize: 1
    keys:
    - sampleKey
    rate: 10`)
	assert.NoError(t, err)

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	start := time.Now()
	count := 0
	for time.Since(start) < time.Duration(2)*time.Second {
		handler.Handle(`{"sampleKey": "a"}`)
		count++
	}
	for i := 0; i < 100; i++ {
		handler.Handle(fmt.Sprintf(`{"sampleKey": "%d"}`, i))
	}
	countsByKey := make(map[string]int)
	for _, ev := range mt.events {
		countsByKey[ev.Data["sampleKey"].(string)]++
	}
	assert.InDelta(t, countsByKey["a"], count/10, 500)
	for i := 0; i < 100; i++ {
		assert.Equal(t, countsByKey[fmt.Sprintf("%d", i)], 1)
	}
}

func TestRequestShaping(t *testing.T) {
	mt := &MockTransmitter{}
	cfg, err := watcherConfigFromYAML(`
---
dataset: kubernetestest
parser: json
processors:
- request_shape:
    field: request
    patterns:
    - /api/:version/:resource
    queryKeys:
    - id`)
	assert.NoError(t, err)
	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")

	testcases := []struct {
		line     string
		expected *event.Event
	}{
		{
			`{"request": "GET /api/v1/users?id=22 HTTP/1.1"}`,
			&event.Event{
				Dataset: "kubernetestest",
				Path:    "/tmp/testpath",
				Data: map[string]interface{}{
					"request":                  "GET /api/v1/users?id=22 HTTP/1.1",
					"request_method":           "GET",
					"request_protocol_version": "HTTP/1.1",
					"request_uri":              "/api/v1/users?id=22",
					"request_path":             "/api/v1/users",
					"request_query":            "id=22",
					"request_shape":            "/api/:version/:resource?id=?",
					"request_path_version":     "v1",
					"request_path_resource":    "users",
					"request_pathshape":        "/api/:version/:resource",
					"request_queryshape":       "id=?",
					"request_query_id":         "22",
				},
			},
		},
		{
			`{"request": "/api/v1/users?id=22"}`,
			&event.Event{
				Dataset: "kubernetestest",
				Path:    "/tmp/testpath",
				Data: map[string]interface{}{
					"request":               "/api/v1/users?id=22",
					"request_uri":           "/api/v1/users?id=22",
					"request_path":          "/api/v1/users",
					"request_query":         "id=22",
					"request_shape":         "/api/:version/:resource?id=?",
					"request_path_version":  "v1",
					"request_path_resource": "users",
					"request_pathshape":     "/api/:version/:resource",
					"request_queryshape":    "id=?",
					"request_query_id":      "22",
				},
			},
		},
	}

	for idx, tc := range testcases {
		handler.Handle(tc.line)
		assert.Equal(t, len(mt.events), idx+1)
		assert.Equal(t, mt.events[idx], tc.expected)
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
