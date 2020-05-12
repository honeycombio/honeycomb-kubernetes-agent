package handlers

// It makes sense for most unit tests to live here, even ones that specifically
// exercise particular processors and such, since they use a common test setup:
// instantiate a LineHandler from a given configuration, pass it some example
// lines, and test that the resultant events match expectations.

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	v1types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

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

type unwrapperType int

const (
	raw unwrapperType = iota
	docker_json
)

type testCase struct {
	config        string
	unwrapperType unwrapperType
	lines         []string
	output        []event.Event
}

func (tc *testCase) check(t *testing.T) {
	cfg, err := watcherConfigFromYAML(tc.config)
	assert.NoError(t, err)
	mt := &MockTransmitter{}

	var unwrapper unwrappers.Unwrapper
	if tc.unwrapperType == raw {
		unwrapper = &unwrappers.RawLogUnwrapper{}
	} else if tc.unwrapperType == docker_json {
		unwrapper = &unwrappers.DockerJSONLogUnwrapper{}
	}

	hf, err := NewLineHandlerFactoryFromConfig(cfg, unwrapper, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	for _, line := range tc.lines {
		handler.Handle(line)
	}
	assert.Equal(t, len(mt.events), len(tc.output))
	for i, out := range tc.output {
		assert.Equal(t, out, *mt.events[i])
	}
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

func TestNginxParsing(t *testing.T) {
	testCases := []*testCase{
		{
			config:        `{"dataset": "kubernetestest", "parser": "nginx"}`,
			unwrapperType: docker_json,
			lines: []string{
				`{"log":"192.168.143.128 - - [10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200 612 \"-\" \"curl/7.38.0\" \"-\"\n","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`,
			},
			output: []event.Event{
				{
					Data: map[string]interface{}{
						"bytes_sent":      int64(612),
						"http_user_agent": "curl/7.38.0",
						"remote_addr":     "192.168.143.128",
						"request":         "GET / HTTP/1.1",
						"status":          int64(200),
						"time_local":      "10/Jul/2017:22:10:25 +0000",
					},
					Dataset:    "kubernetestest",
					Path:       "/tmp/testpath",
					Timestamp:  time.Date(2017, 07, 10, 22, 10, 25, 569584932, time.UTC),
					RawMessage: "192.168.143.128 - - [10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200 612 \"-\" \"curl/7.38.0\" \"-\"",
				},
			},
		},
		{
			config: `{
				"dataset": "kubernetestest",
				"parser": { "name": "nginx", "options": { "log_format": "envoy" } },
				"processors": ["timefield": { "field": "timestamp" }]
			}`,
			unwrapperType: docker_json,
			lines: []string{
				`{"log":"[2016-04-15T20:17:00.310Z] \"POST /api/v1/locations HTTP/2\" 204 - 154 0 226 100 \"10.0.35.28\" \"nsq2http\" \"cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2\" \"locations\" \"tcp://10.0.2.1:80\"\n","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`,
			},
			output: []event.Event{
				{
					Data: map[string]interface{}{
						"request":                       "POST /api/v1/locations HTTP/2",
						"status_code":                   int64(204),
						"bytes_received":                int64(154),
						"bytes_sent":                    int64(0),
						"duration":                      int64(226),
						"x_envoy_upstream_service_time": int64(100),
						"x_forwarded_for":               "10.0.35.28",
						"user_agent":                    "nsq2http",
						"x_request_id":                  "cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2",
						"authority":                     "locations",
						"upstream_host":                 "tcp://10.0.2.1:80",
					},
					Dataset:    "kubernetestest",
					Path:       "/tmp/testpath",
					Timestamp:  time.Date(2016, 4, 15, 20, 17, 0, 310000000, time.UTC),
					RawMessage: "[2016-04-15T20:17:00.310Z] \"POST /api/v1/locations HTTP/2\" 204 - 154 0 226 100 \"10.0.35.28\" \"nsq2http\" \"cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2\" \"locations\" \"tcp://10.0.2.1:80\"",
				},
			},
		},
		{
			config: `{
				"dataset": "kubernetestest",
				"parser": { "name": "envoy" },
				"processors": ["timefield": { "field": "timestamp" }]
			}`,
			unwrapperType: docker_json,
			lines: []string{
				`{"log":"[2016-04-15T20:17:00.310Z] \"POST /api/v1/locations HTTP/2\" 204 - 154 0 226 100 \"10.0.35.28\" \"nsq2http\" \"cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2\" \"locations\" \"tcp://10.0.2.1:80\"\n","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`,
			},
			output: []event.Event{
				{
					Data: map[string]interface{}{
						"request":                       "POST /api/v1/locations HTTP/2",
						"status_code":                   int64(204),
						"bytes_received":                int64(154),
						"bytes_sent":                    int64(0),
						"duration":                      int64(226),
						"x_envoy_upstream_service_time": int64(100),
						"x_forwarded_for":               "10.0.35.28",
						"user_agent":                    "nsq2http",
						"x_request_id":                  "cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2",
						"authority":                     "locations",
						"upstream_host":                 "tcp://10.0.2.1:80",
					},
					Dataset:    "kubernetestest",
					Path:       "/tmp/testpath",
					Timestamp:  time.Date(2016, 4, 15, 20, 17, 0, 310000000, time.UTC),
					RawMessage: "[2016-04-15T20:17:00.310Z] \"POST /api/v1/locations HTTP/2\" 204 - 154 0 226 100 \"10.0.35.28\" \"nsq2http\" \"cc21d9b0-cf5c-432b-8c7e-98aeb7988cd2\" \"locations\" \"tcp://10.0.2.1:80\"",
				},
			},
		},
		{
			config: `{
				"dataset": "kubernetestest",
				"parser": { "name": "nginx-ingress" },
				"processors":
			}`,
			unwrapperType: docker_json,
			lines: []string{
				`{"log":"10.0.0.1 - [10.0.0.1] - - [14/Jun/2018:18:20:48 +0000] \"GET /api/v1/users?id=22 HTTP/1.1\" 200 1198 \"-\" \"curl\" 536 0.165 [api-22] 10.0.0.2:10001 1202 0.165 200 abcd\n","stream":"stdout","time":"2017-07-10T22:10:25Z"}`,
			},
			output: []event.Event{
				{
					Data: map[string]interface{}{
						"the_real_ip":              "10.0.0.1",
						"time_local":               "14/Jun/2018:18:20:48 +0000",
						"request":                  "GET /api/v1/users?id=22 HTTP/1.1",
						"status":                   int64(200),
						"body_bytes_sent":          int64(1198),
						"http_user_agent":          "curl",
						"request_length":           int64(536),
						"request_time":             0.165,
						"proxy_upstream_name":      "api-22",
						"upstream_addr":            "10.0.0.2:10001",
						"upstream_response_length": int64(1202),
						"upstream_response_time":   0.165,
						"upstream_status":          int64(200),
						"req_id":                   "abcd",
					},
					Dataset:    "kubernetestest",
					Path:       "/tmp/testpath",
					Timestamp:  time.Date(2017, 7, 10, 22, 10, 25, 0, time.UTC),
					RawMessage: "10.0.0.1 - [10.0.0.1] - - [14/Jun/2018:18:20:48 +0000] \"GET /api/v1/users?id=22 HTTP/1.1\" 200 1198 \"-\" \"curl\" 536 0.165 [api-22] 10.0.0.2:10001 1202 0.165 200 abcd",
				},
			},
		},
		{
			config: `{
				"dataset": "kubernetestest",
				"parser": {
					"name": "nginx",
					"options": { "log_format": "[$time_local] \"$request\" $status" }
				}
			}`,
			unwrapperType: docker_json,
			lines: []string{
				`{"log":"[10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200","stream":"stdout","time":"2017-07-10T22:10:25.569584932Z"}`,
			},
			output: []event.Event{
				{
					Data: map[string]interface{}{
						"request":    "GET / HTTP/1.1",
						"status":     int64(200),
						"time_local": "10/Jul/2017:22:10:25 +0000",
					},
					Dataset:    "kubernetestest",
					Path:       "/tmp/testpath",
					Timestamp:  time.Date(2017, 07, 10, 22, 10, 25, 569584932, time.UTC),
					RawMessage: "[10/Jul/2017:22:10:25 +0000] \"GET / HTTP/1.1\" 200",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), tc.check)
	}
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
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		Timestamp:  time.Date(2017, 7, 10, 22, 10, 25, 569584932, time.UTC),
		RawMessage: `W0720 00:15:01.592300       5 controller.go:386] Resetting endpoints for master service`,
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
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		Timestamp:  time.Date(2017, 7, 2, 22, 10, 25, 569534932, time.UTC),
		RawMessage: `44:C 09 Aug 23:12:19.127 * RDB: 0 MB of memory used by copy-on-write`,
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestKeyvalParsing(t *testing.T) {
	tc := testCase{
		config: `
---
dataset: kubernetestest
parser:
  name: keyval
  options:
    prefixRegex: "(?P<timestamp>[0-9\\.:TZ\\-]+) AUDIT: "
processors:
  - timefield:
      field: timestamp
`,
		lines: []string{
			`2017-08-25T04:40:49.965969122Z AUDIT: id="8e2ad929-8da1-4ce5-8776-751421157483" ip="172.20.67.135" method="GET" user="system:serviceaccount:tectonic-system:prometheus-operator" groups="\"system:serviceaccounts\",\"system:serviceaccounts:tectonic-system\",\"system:authenticated\"" as="<self>" asgroups="<lookup>" namespace="tectonic-system" uri="/api/v1/namespaces/tectonic-system/secrets/prometheus-k8s"`,
		},
		output: []event.Event{
			event.Event{
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
				Dataset:    "kubernetestest",
				Path:       "/tmp/testpath",
				Timestamp:  time.Date(2017, 8, 25, 4, 40, 49, 965969122, time.UTC),
				RawMessage: `2017-08-25T04:40:49.965969122Z AUDIT: id="8e2ad929-8da1-4ce5-8776-751421157483" ip="172.20.67.135" method="GET" user="system:serviceaccount:tectonic-system:prometheus-operator" groups="\"system:serviceaccounts\",\"system:serviceaccounts:tectonic-system\",\"system:authenticated\"" as="<self>" asgroups="<lookup>" namespace="tectonic-system" uri="/api/v1/namespaces/tectonic-system/secrets/prometheus-k8s"`,
			},
		},
	}

	tc.check(t)
}

func TestKubernetesAuditLogHandling(t *testing.T) {
	tc := &testCase{
		config: `{
			"dataset": "kubernetestest",
			"parser": "audit",
			"processors": ["timefield": {"field": "timestamp"}]
		}`,
		unwrapperType: raw,
		lines: []string{
			`2017-03-21T03:57:09.106841886+04:00 AUDIT: id="c939d2a7-1c37-4ef1-b2f7-4ba9b1e43b53" ip="127.0.0.1" method="GET" user="admin" groups="\"system:masters\",\"system:authenticated\"" as="<self>" asgroups="<lookup>" namespace="default" uri="/api/v1/namespaces/default/pods"`,
			`2017-03-21T03:57:09.108403639+04:00 AUDIT: id="c939d2a7-1c37-4ef1-b2f7-4ba9b1e43b53" response="200"`,
		},
		output: []event.Event{
			event.Event{
				Data: map[string]interface{}{
					"id":        "c939d2a7-1c37-4ef1-b2f7-4ba9b1e43b53",
					"ip":        "127.0.0.1",
					"method":    "GET",
					"user":      "admin",
					"groups":    "\"system:masters\",\"system:authenticated\"",
					"as":        "<self>",
					"asgroups":  "<lookup>",
					"namespace": "default",
					"uri":       "/api/v1/namespaces/default/pods",
					"response":  200,
				},
				Dataset:   "kubernetestest",
				Path:      "/tmp/testpath",
				Timestamp: time.Date(2017, 3, 21, 3, 57, 9, 106841886, time.FixedZone("", 4*3600)),
				// Passing the "raw message" for a multi-line parser isn't perfect, but it's the best
				// we're willing to do for now.
				RawMessage: `2017-03-21T03:57:09.108403639+04:00 AUDIT: id="c939d2a7-1c37-4ef1-b2f7-4ba9b1e43b53" response="200"`,
			},
		},
	}

	tc.check(t)
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
		Data:       map[string]interface{}{"dontdrop": "b"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"todrop": "a", "dontdrop": "b"}`,
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestScrubField(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "json"},
		Processors: []map[string]map[string]interface{}{
			map[string]map[string]interface{}{
				"scrub_field": map[string]interface{}{"field": "toscrub"},
			},
		},
	}

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"toscrub": "a", "dontscrub": "b"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data:       map[string]interface{}{"toscrub": "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", "dontscrub": "b"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"toscrub": "a", "dontscrub": "b"}`,
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestRenameField(t *testing.T) {
	mt := &MockTransmitter{}
	cfg := &config.WatcherConfig{
		Dataset: "kubernetestest",
		Parser:  &config.ParserConfig{Name: "json"},
		Processors: []map[string]map[string]interface{}{
			map[string]map[string]interface{}{
				"rename_field": map[string]interface{}{"original": "boring", "new": "interesting"},
			},
		},
	}

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"boring": "field", "another": "field"}`)
	assert.Equal(t, len(mt.events), 1)
	expected := &event.Event{
		Data:       map[string]interface{}{"interesting": "field", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"boring": "field", "another": "field"}`,
	}
	assert.Equal(t, mt.events[0], expected)
}

func TestEventRouter(t *testing.T) {
	mt := &MockTransmitter{}

	cfg, err := watcherConfigFromYAML(`
dataset: kubernetestest
parser: json
processors:
- route_event:
    field: host
    routes:
      - value: api
        dataset: api
      - value: www
        dataset: not-api`)
	assert.NoError(t, err)

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"host": "default", "another": "field"}`)
	handler.Handle(`{"host": "api", "another": "field"}`)
	handler.Handle(`{"host": "www", "another": "field"}`)
	assert.Equal(t, len(mt.events), 3)
	expected0 := &event.Event{
		Data:       map[string]interface{}{"host": "default", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"host": "default", "another": "field"}`,
	}
	expected1 := &event.Event{
		Data:       map[string]interface{}{"host": "api", "another": "field"},
		Dataset:    "api",
		Path:       "/tmp/testpath",
		RawMessage: `{"host": "api", "another": "field"}`,
	}
	expected2 := &event.Event{
		Data:       map[string]interface{}{"host": "www", "another": "field"},
		Dataset:    "not-api",
		Path:       "/tmp/testpath",
		RawMessage: `{"host": "www", "another": "field"}`,
	}
	assert.Equal(t, mt.events[0], expected0)
	assert.Equal(t, mt.events[1], expected1)
	assert.Equal(t, mt.events[2], expected2)
}

func TestEventDropper(t *testing.T) {
	mt := &MockTransmitter{}

	cfg, err := watcherConfigFromYAML(`
dataset: kubernetestest
parser: json
processors:
- drop_event:
    field: service
    values:
      - dropthis`)
	assert.NoError(t, err)

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")

	handler.Handle(`{"service": "dropthis", "another": "field"}`)
	handler.Handle(`{"service": "keepme", "another": "field"}`)
	handler.Handle(`{"no_service": "keepme", "another": "field"}`)
	assert.Equal(t, len(mt.events), 2)
	expected0 := &event.Event{
		Data:       map[string]interface{}{"service": "keepme", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"service": "keepme", "another": "field"}`,
	}
	expected1 := &event.Event{
		Data:       map[string]interface{}{"no_service": "keepme", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"no_service": "keepme", "another": "field"}`,
	}
	assert.Equal(t, mt.events[0], expected0)
	assert.Equal(t, mt.events[1], expected1)
}

func TestEventKeeper(t *testing.T) {
	mt := &MockTransmitter{}

	cfg, err := watcherConfigFromYAML(`
dataset: kubernetestest
parser: json
processors:
- keep_event:
    field: service
    values:
      - keepme`)
	assert.NoError(t, err)

	hf, err := NewLineHandlerFactoryFromConfig(cfg, &unwrappers.RawLogUnwrapper{}, mt)
	assert.NoError(t, err)
	handler := hf.New("/tmp/testpath")
	handler.Handle(`{"service": "dropthis", "another": "field"}`)
	handler.Handle(`{"service": "keepme", "another": "field"}`)
	handler.Handle(`{"no_service": "keepme", "another": "field"}`)
	assert.Equal(t, len(mt.events), 2)
	expected0 := &event.Event{
		Data:       map[string]interface{}{"service": "keepme", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"service": "keepme", "another": "field"}`,
	}
	expected1 := &event.Event{
		Data:       map[string]interface{}{"no_service": "keepme", "another": "field"},
		Dataset:    "kubernetestest",
		Path:       "/tmp/testpath",
		RawMessage: `{"no_service": "keepme", "another": "field"}`,
	}
	assert.Equal(t, mt.events[0], expected0)
	assert.Equal(t, mt.events[1], expected1)
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
				RawMessage: `{"request": "GET /api/v1/users?id=22 HTTP/1.1"}`,
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
				RawMessage: `{"request": "/api/v1/users?id=22"}`,
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

func (mp *mockPodGetter) Pods() chan *v1.Pod { return nil }

func (mp *mockPodGetter) DeletedPods() chan types.UID { return nil }

func (mp *mockPodGetter) Get(uid types.UID) (*v1.Pod, bool) {
	return &v1.Pod{
		ObjectMeta: v1types.ObjectMeta{
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
