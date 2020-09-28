package metrics

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics/mock"
	"github.com/sirupsen/logrus"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateMetricsData(t *testing.T) {
	rc := &metrics_mock.MockRestClient{}
	statsProvider := kubelet.NewStatsProvider(rc)
	summary, _ := statsProvider.StatsSummary()
	metadataProvider := kubelet.NewMetadataProvider(rc)
	podsMetadata, _ := metadataProvider.Pods()
	metadata := NewMetadata(podsMetadata, nil, logrus.StandardLogger())

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	require.Equal(t, 42, len(p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{
		"node":      true,
		"pod":       true,
		"container": true,
		"volume":    true,
	})))

	require.Equal(t, 12, len(p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{
		"node":      true,
		"pod":       true,
		"container": false,
		"volume":    false,
	})))
}
