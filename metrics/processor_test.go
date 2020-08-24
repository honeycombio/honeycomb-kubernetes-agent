package metrics

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRestClient struct {
}

func (f fakeRestClient) StatsSummary() ([]byte, error) {
	return ioutil.ReadFile("../testdata/stats-summary.json")
}

func (f fakeRestClient) Pods() ([]byte, error) {
	return ioutil.ReadFile("../testdata/pods.json")
}

func TestMetricAccumulator(t *testing.T) {
	rc := &fakeRestClient{}
	statsProvider := kubelet.NewStatsProvider(rc)
	summary, _ := statsProvider.StatsSummary()
	metadataProvider := kubelet.NewMetadataProvider(rc)
	podsMetadata, _ := metadataProvider.Pods()
	metadata := NewMetadata(podsMetadata, nil, logrus.StandardLogger())

	// Disable all groups
	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())
	require.Equal(t, 28, len(p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{
		"node":      true,
		"pod":       true,
		"container": true,
		"volume":    true,
	})))
}
