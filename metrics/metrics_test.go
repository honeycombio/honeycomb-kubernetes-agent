package metrics

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	metrics_mock "github.com/honeycombio/honeycomb-kubernetes-agent/metrics/mock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createMockSourceAssets(isSum bool, isMeta bool, omitLabels []OmitLabel) (*stats.Summary, *Metadata) {
	rc := &metrics_mock.MockRestClient{}

	var summary *stats.Summary
	if isSum {
		statsProvider := kubelet.NewStatsProvider(rc)
		summary, _ = statsProvider.StatsSummary()
	}

	var metadata *Metadata
	if isMeta {
		metadataProvider := kubelet.NewMetadataProvider(rc)
		podsMetadata, _ := metadataProvider.Pods()
		metadata = NewMetadata(podsMetadata, omitLabels, logrus.StandardLogger())
	}

	return summary, metadata
}

func createMockAccumulator(metadata *Metadata, mg map[MetricGroup]bool) *MetricDataAccumulator {

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	return &MetricDataAccumulator{
		metadata:              metadata,
		metricGroupsToCollect: mg,
		mp:                    p,
		time:                  time.Now(),
		logger:                logrus.StandardLogger(),
	}
}

func getPodStatsByUID(pods []stats.PodStats, uid string) stats.PodStats {
	for _, pod := range pods {
		if pod.PodRef.UID == uid {
			return pod
		}
	}
	return stats.PodStats{}
}

func getPodMetric(resMetrics []*ResourceMetrics, res *Resource, name string) *Metric {
	for _, rm := range resMetrics {
		if rm.Resource.Name == res.Name {
			for k, v := range rm.Metrics {
				if k == name {
					return v
				}
			}
			return nil
		}
	}
	return nil
}

func TestGetLabels(t *testing.T) {

	_, md := createMockSourceAssets(false, true, []OmitLabel{
		"controller-revision-hash",
		"pod-template-generation",
	})
	pmd, _ := md.GetPodMetadataByUid("5c69ffd4-73d1-47df-87d7-08ae860e75ae")

	labels := pmd.GetLabels()

	assert.Equal(t, 3, len(labels))

	assert.Equal(t, "honeycomb", labels["app.kubernetes.io/name"])
}

func TestGetCpuLimit(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	limit := pmd.GetCpuLimit()
	assert.Equal(t, 0.1, limit)

	pmd, _ = md.GetPodMetadataByUid("c0beb6af-9b87-4e0d-a143-936c9ab7f63b")
	limit = pmd.GetCpuLimit()
	assert.Equal(t, float64(0), limit)
}

func TestGetCpuLimitForContainer(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	limit := pmd.GetCpuLimitForContainer("speaker")
	assert.Equal(t, 0.1, limit)
}

func TestGetMemoryLimit(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	limit := pmd.GetMemoryLimit()
	assert.Equal(t, 100*math.Pow(2, 20), limit)

	pmd, _ = md.GetPodMetadataByUid("c0beb6af-9b87-4e0d-a143-936c9ab7f63b")
	limit = pmd.GetMemoryLimit()
	assert.Equal(t, float64(0), limit)
}

func TestGetMemoryLimitForContainer(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	limit := pmd.GetMemoryLimitForContainer("speaker")
	assert.Equal(t, 100*math.Pow(2, 20), limit)
}

func TestGetStatus(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	status := pmd.GetStatus()
	assert.Equal(t, "Running", status[StatusPhase])
}

func TestGetStatusForContainer(t *testing.T) {
	_, md := createMockSourceAssets(false, true, nil)

	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	status := pmd.GetStatusForContainer("speaker")
	assert.Equal(t, int32(54), status[StatusRestartCount])
	assert.Equal(t, true, status[StatusReady])
	assert.Equal(t, "running", status[StatusState])
}

func TestGenerateMetricsData(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	assert.Equal(t, 42, len(p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{
		"node":      true,
		"pod":       true,
		"container": true,
		"volume":    true,
	})))

	assert.Equal(t, 12, len(p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{
		"node":      true,
		"pod":       true,
		"container": false,
		"volume":    false,
	})))
}

func TestCpuMetrics(t *testing.T) {
	summary, _ := createMockSourceAssets(true, false, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.CpuMetrics(summary.Pods[0].CPU, 0.2) //JAMIETEST

	require.Equal(t, 2, len(metrics))

	assert.Equal(t, 0.015338365, metrics[MeasureCpuUsage].GetValue(), "CPU Usage")
	assert.InDelta(t, 0.000001, 7.6691825, metrics[MeasureCpuUtilization].GetValue(), "CPU Utilization")
}

func TestCpuMetricsOptional(t *testing.T) {
	summary, _ := createMockSourceAssets(true, false, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.CpuMetrics(summary.Pods[1].CPU, 0) //JAMIETEST

	require.Equal(t, 0, len(metrics))

	assert.Equal(t, 0, metrics[MeasureCpuUsage].GetValue(), "CPU Usage")
	assert.Equal(t, 0, metrics[MeasureCpuUtilization].GetValue(), "CPU Utilization")
}

func TestMemMetrics(t *testing.T) {
	summary, _ := createMockSourceAssets(true, false, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.MemMetrics(summary.Pods[0].Memory, 2147483648)

	require.Equal(t, 7, len(metrics))

	assert.Equal(t, float64(2143281152), metrics[MeasureMemoryUsage].GetValue(), "Memory Usage")
	assert.InDelta(t, 0.000001, 99.804300603, metrics[MeasureMemoryUtilization].GetValue(), "Memory Utilization")
	assert.Equal(t, float64(23191552), metrics[MeasureMemoryRSS].GetValue(), "Memory RSS")
}

func TestNetworkMetrics(t *testing.T) {
	summary, _ := createMockSourceAssets(true, false, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.NetworkMetrics(summary.Pods[0].Network)

	require.Equal(t, 4, len(metrics))

	assert.Equal(t, float64(150353577), metrics[MeasureNetworkBytesReceive].GetValue(), "Network Bytes Received")
	assert.Equal(t, float64(0), metrics[MeasureNetworkErrorsReceive].GetValue(), "Network Errors Received")
}

func TestVolumeMetrics(t *testing.T) {
	summary, _ := createMockSourceAssets(true, false, nil)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.VolumeMetrics(summary.Pods[0].VolumeStats[0])

	require.Equal(t, 6, len(metrics))

	assert.Equal(t, float64(4182040576), metrics[MeasureVolumeAvailable].GetValue(), "Volume Available")
	assert.Equal(t, float64(12288), metrics[MeasureVolumeUsed].GetValue(), "Volume Used")
	assert.Equal(t, float64(1021009), metrics[MeasureVolumeInodesTotal].GetValue(), "Volume Inodes Total")
}

func TestCounterMetrics(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)

	podStats := getPodStatsByUID(summary.Pods, "5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	node := getNodeResource(summary.Node)
	pod := getPodResource(node, podStats, metadata)

	p := NewMetricsProcessor(10*time.Second, logrus.StandardLogger())

	metrics := p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{"pod": true})

	m := getPodMetric(metrics, pod, MeasureCpuUsage)
	val := p.GetCounterRate(pod, MeasureCpuUsage, m)
	assert.Equal(t, float64(0), val)

	cpuUsage := float64(1500000000) / 1000000000
	nm := &Metric{
		Type:       MetricTypeFloat,
		FloatValue: &cpuUsage,
	}
	pod.Timestamp = pod.Timestamp.Add(10 * time.Second)
	val = p.GetCounterRate(pod, MeasureCpuUsage, nm)
	assert.InDelta(t, 0.150, val, 0.01)
}

func TestExpiredCounterMetrics(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)

	podStats := getPodStatsByUID(summary.Pods, "5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	node := getNodeResource(summary.Node)
	pod := getPodResource(node, podStats, metadata)

	// short expiration time to allow test to run faster
	p := NewMetricsProcessor(1*time.Second, logrus.StandardLogger())

	metrics := p.GenerateMetricsData(summary, metadata, map[MetricGroup]bool{"pod": true})

	m := getPodMetric(metrics, pod, MeasureCpuUsage)
	val := p.GetCounterRate(pod, MeasureCpuUsage, m)
	assert.Equal(t, float64(0), val)

	cpuUsage := float64(1500000000) / 1000000000
	nm := &Metric{
		Type:       MetricTypeFloat,
		FloatValue: &cpuUsage,
	}

	time.Sleep(7 * time.Second)

	pod.Timestamp = time.Now()
	val = p.GetCounterRate(pod, MeasureCpuUsage, nm)
	assert.Equal(t, float64(0), val)
}

func TestNodeStats(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)
	acc := createMockAccumulator(metadata, ValidMetricGroups)

	node := getNodeResource(summary.Node)

	acc.nodeStats(node, summary.Node)
	data := acc.Data[0]

	assert.Equal(t, "duckboat-01", data.Resource.Name)
	assert.Equal(t, 0, len(data.Resource.Status))
	assert.Equal(t, 1, len(data.Resource.Labels))
	assert.Equal(t, 17, len(data.Metrics))
	assert.Equal(t, float64(388954406)/1000000000, data.Metrics[MeasureCpuUsage].GetValue())

}

func TestPodStats(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)
	acc := createMockAccumulator(metadata, ValidMetricGroups)

	podStats := getPodStatsByUID(summary.Pods, "5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	node := getNodeResource(summary.Node)
	pod := getPodResource(node, podStats, metadata)

	acc.podStats(pod, podStats)
	data := acc.Data[0]

	assert.Equal(t, "speaker-cpxhz", data.Resource.Name)
	assert.Equal(t, 4, len(data.Resource.Status))
	assert.Equal(t, 8, len(data.Resource.Labels))
	assert.Equal(t, "metallb", data.Resource.Labels[PrefixLabel+"app"])
	assert.Equal(t, "metallb-system", data.Resource.Labels[LabelNamespaceName])
	assert.Equal(t, 17, len(data.Metrics))
	assert.Equal(t, float64(7919180)/1000000000, data.Metrics[MeasureCpuUsage].GetValue())

}

func TestContainerStats(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)
	acc := createMockAccumulator(metadata, ValidMetricGroups)

	podStats := getPodStatsByUID(summary.Pods, "5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	node := getNodeResource(summary.Node)
	pod := getPodResource(node, podStats, metadata)

	acc.containerStats(pod, podStats.Containers[0])
	data := acc.Data[0]

	assert.Equal(t, "speaker", data.Resource.Name)
	assert.Equal(t, 4, len(data.Resource.Status))
	assert.Equal(t, "running", data.Resource.Status[StatusState])
	assert.Equal(t, 9, len(data.Resource.Labels))
	assert.Equal(t, "metallb", data.Resource.Labels[PrefixLabel+"app"])
	assert.Equal(t, "metallb-system", data.Resource.Labels[LabelNamespaceName])
	assert.Equal(t, 13, len(data.Metrics))
	assert.Equal(t, float64(13107067)/1000000000, data.Metrics[MeasureCpuUsage].GetValue())

}

func TestVolumeStats(t *testing.T) {
	summary, metadata := createMockSourceAssets(true, true, nil)
	acc := createMockAccumulator(metadata, ValidMetricGroups)

	podStats := getPodStatsByUID(summary.Pods, "5997ad9b-1d2a-43cf-ab57-a98d8796dc34")
	node := getNodeResource(summary.Node)
	pod := getPodResource(node, podStats, metadata)

	acc.volumeStats(pod, podStats.VolumeStats[0])
	data := acc.Data[0]

	assert.Equal(t, "speaker-token-kpzds", data.Resource.Name)
	assert.Equal(t, 0, len(data.Resource.Status))
	assert.Equal(t, 9, len(data.Resource.Labels))
	assert.Equal(t, "metallb", data.Resource.Labels[PrefixLabel+"app"])
	assert.Equal(t, "metallb-system", data.Resource.Labels[LabelNamespaceName])
	assert.Equal(t, 6, len(data.Metrics))

}
