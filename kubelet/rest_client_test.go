// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestClient(t *testing.T) {
	rest := NewRestClient(&fakeClient{})
	resp, _ := rest.StatsSummary()
	require.Equal(t, "/stats/summary", string(resp))
	resp, _ = rest.Pods()
	require.Equal(t, "/pods", string(resp))
}

var _ Client = (*fakeClient)(nil)

type fakeClient struct{}

func (f *fakeClient) Get(path string) ([]byte, error) {
	return []byte(path), nil
}
