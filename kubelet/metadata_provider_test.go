// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

import (
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRestClient struct {
	fail        bool
	invalidJSON bool
}

func (f testRestClient) StatsSummary() ([]byte, error) {
	return []byte{}, nil
}

func (f testRestClient) Pods() ([]byte, error) {
	if f.fail {
		return []byte{}, errors.New("failed")
	}
	if f.invalidJSON {
		return []byte("wrong-json-body"), nil
	}

	return ioutil.ReadFile("../testdata/pods.json")
}

func TestPods(t *testing.T) {
	tests := []struct {
		name      string
		client    RestClient
		wantError string
	}{
		{
			name:      "success",
			client:    &testRestClient{},
			wantError: "",
		},
		{
			name:      "failure",
			client:    &testRestClient{fail: true},
			wantError: "failed",
		},
		{
			name:      "invalid-json",
			client:    &testRestClient{invalidJSON: true},
			wantError: "invalid character 'w' looking for beginning of value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataProvider := NewMetadataProvider(tt.client)
			podsMetadata, err := metadataProvider.Pods()
			if tt.wantError == "" {
				require.NoError(t, err)
				require.Less(t, 0, len(podsMetadata.Items))
			} else {
				if err == nil {
					assert.Fail(t, "err is nil", tt.name)
				} else {
					assert.Equal(t, tt.wantError, err.Error())
				}
			}
		})
	}
}
