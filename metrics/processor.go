package metrics

import (
	"github.com/sirupsen/logrus"

	"time"

	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

type Processor struct {
	counterCache *Cache
	logger       *logrus.Logger
}

func NewMetricsProcessor(fetchInterval time.Duration, logger *logrus.Logger) *Processor {
	timeout := fetchInterval * (5 + 1)
	cache := NewCache(timeout)
	return &Processor{
		counterCache: cache,
		logger:       logger,
	}

}

func (p *Processor) GenerateMetricsData(summary *stats.Summary, metadata *Metadata, metricGroupsToCollect map[MetricGroup]bool) []*ResourceMetrics {

	acc := &MetricDataAccumulator{
		metadata:              metadata,
		metricGroupsToCollect: metricGroupsToCollect,
		mp:                    p,
		time:                  time.Now(),
		logger:                p.logger,
	}

	nodeResource := getNodeResource(summary.Node)
	acc.nodeStats(nodeResource, summary.Node)
	for _, podStats := range summary.Pods {
		// propagate the pod resource down to the container
		podResource := getPodResource(nodeResource, podStats, metadata)
		acc.podStats(podResource, podStats)
		for _, containerStats := range podStats.Containers {
			acc.containerStats(podResource, containerStats)
		}

		for _, volumeStats := range podStats.VolumeStats {
			acc.volumeStats(podResource, volumeStats)
		}
	}

	return acc.Data
}

func (p *Processor) GetCounterRate(res *Resource, name string, newMetric *Metric) float64 {

	key := res.Name + "." + name

	var rate float64
	if counter, ok := p.counterCache.Get(key); ok {
		// CounterValue exists, compute per second rate

		delta := newMetric.GetValue() - counter.Value.GetValue()
		secs := res.Timestamp.Unix() - counter.Timestamp.Unix()

		if secs == 0 {
			rate = 0
		} else {
			rate = delta / float64(secs)
		}

	} else {
		// CounterValue doesn't already exists, return 0
		rate = 0
	}

	// Store new counter value
	newCounter := &CounterValue{
		Key:       key,
		Value:     newMetric,
		Timestamp: res.Timestamp,
	}
	p.counterCache.Set(key, newCounter)

	return rate
}

func (p *Processor) UptimeMetrics(startTime time.Time) Metrics {
	var uptime uint64
	if startTime.IsZero() {
		uptime = 0
	} else {
		uptime = uint64(time.Since(startTime).Nanoseconds() / time.Millisecond.Nanoseconds())
	}

	return Metrics{
		MeasureUptime: &Metric{Type: MetricTypeInt, IntValue: &uptime},
	}

}

func (p *Processor) CpuMetrics(s *stats.CPUStats, limit float64) Metrics {
	var cpuUsage float64
	nanoCores := s.UsageNanoCores
	if nanoCores == nil {
		cpuUsage = 0
	} else {
		cpuUsage = float64(*nanoCores) / 1000000000
	}

	cpuUtilization := float64(0)
	if limit > 0 {
		cpuUtilization = (cpuUsage / limit) * 100
	}

	return Metrics{
		MeasureCpuUsage:       &Metric{Type: MetricTypeFloat, FloatValue: &cpuUsage},
		MeasureCpuUtilization: &Metric{Type: MetricTypeFloat, FloatValue: &cpuUtilization},
	}
}

func (p *Processor) MemMetrics(s *stats.MemoryStats, limit float64) Metrics {
	var utilization float64
	if limit > 0 {
		usage := float64(*s.UsageBytes)
		utilization = (usage / limit) * 100
	}
	return Metrics{
		MeasureMemoryAvailable:       &Metric{Type: MetricTypeInt, IntValue: s.AvailableBytes},
		MeasureMemoryUsage:           &Metric{Type: MetricTypeInt, IntValue: s.UsageBytes},
		MeasureMemoryUtilization:     &Metric{Type: MetricTypeFloat, FloatValue: &utilization},
		MeasureMemoryRSS:             &Metric{Type: MetricTypeInt, IntValue: s.RSSBytes},
		MeasureMemoryWorkingSet:      &Metric{Type: MetricTypeInt, IntValue: s.WorkingSetBytes},
		MeasureMemoryPageFaults:      &Metric{Type: MetricTypeInt, IntValue: s.PageFaults},
		MeasureMemoryMajorPageFaults: &Metric{Type: MetricTypeInt, IntValue: s.MajorPageFaults},
	}
}

func (p *Processor) FsMetrics(s *stats.FsStats) Metrics {
	if s == nil {
		return Metrics{}
	}

	return Metrics{
		MeasureFilesystemAvailable: &Metric{Type: MetricTypeInt, IntValue: s.AvailableBytes},
		MeasureFilesystemCapacity:  &Metric{Type: MetricTypeInt, IntValue: s.CapacityBytes},
		MeasureFilesystemUsage:     &Metric{Type: MetricTypeInt, IntValue: s.UsedBytes},
	}
}

func (p *Processor) NetworkMetrics(s *stats.NetworkStats) Metrics {
	if s == nil {
		return Metrics{}
	}

	return Metrics{
		MeasureNetworkBytesReceive:  &Metric{Type: MetricTypeInt, IsCounter: true, IntValue: s.RxBytes},
		MeasureNetworkBytesSend:     &Metric{Type: MetricTypeInt, IsCounter: true, IntValue: s.TxBytes},
		MeasureNetworkErrorsReceive: &Metric{Type: MetricTypeInt, IsCounter: true, IntValue: s.RxErrors},
		MeasureNetworkErrorsSend:    &Metric{Type: MetricTypeInt, IsCounter: true, IntValue: s.TxErrors},
	}
}

func (p *Processor) VolumeMetrics(s stats.VolumeStats) Metrics {
	return Metrics{
		MeasureVolumeAvailable:   &Metric{Type: MetricTypeInt, IntValue: s.AvailableBytes},
		MeasureVolumeCapacity:    &Metric{Type: MetricTypeInt, IntValue: s.CapacityBytes},
		MeasureVolumeInodesTotal: &Metric{Type: MetricTypeInt, IntValue: s.Inodes},
		MeasureVolumeInodesFree:  &Metric{Type: MetricTypeInt, IntValue: s.InodesFree},
		MeasureVolumeInodesUsed:  &Metric{Type: MetricTypeInt, IntValue: s.InodesUsed},
	}
}
