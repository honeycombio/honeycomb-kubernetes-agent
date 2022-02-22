package metrics

import (
	"time"

	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type Processor struct {
	counterCache *Cache
}

func NewMetricsProcessor(fetchInterval time.Duration) *Processor {
	timeout := fetchInterval * (5 + 1)
	cache := NewCache(timeout)
	return &Processor{
		counterCache: cache,
	}

}

func (p *Processor) GenerateMetricsData(summary *stats.Summary, metadata *Metadata, metricGroupsToCollect map[MetricGroup]bool) []*ResourceMetrics {

	acc := &MetricDataAccumulator{
		metadata:              metadata,
		metricGroupsToCollect: metricGroupsToCollect,
		mp:                    p,
		time:                  time.Now(),
	}

	nodeResource := getNodeResource(summary.Node, metadata)
	acc.nodeStats(nodeResource, summary.Node)

	for _, podStats := range summary.Pods {

		podResource := getPodResource(nodeResource, podStats, metadata)
		acc.podStats(podResource, podStats)

		// Container stats rely on podResource for meta
		for _, containerStats := range podStats.Containers {
			acc.containerStats(podResource, containerStats)
		}

		// Volume stats rely on podResource for meta
		for _, volumeStats := range podStats.VolumeStats {
			acc.volumeStats(podResource, volumeStats)
		}
	}

	return acc.Data
}

func (p *Processor) GetCounterDelta(res *Resource, name string, newMetric *Metric) float64 {
	key := p.getCounterKey(res, name)

	var delta float64
	if counter, ok := p.counterCache.Get(key); ok {
		// CounterValue exists, compute delta
		delta = newMetric.GetValue() - counter.Value.GetValue()
	}

	// Store new counter value
	p.counterCache.Set(key, &CounterValue{
		Key:       key,
		Value:     newMetric,
		Timestamp: res.Timestamp,
	})

	return delta
}

func (p *Processor) GetCounterRate(res *Resource, name string, newMetric *Metric) float64 {
	key := p.getCounterKey(res, name)

	var rate float64
	if counter, ok := p.counterCache.Get(key); ok {
		// CounterValue exists, compute per second rate

		delta := newMetric.GetValue() - counter.Value.GetValue()
		secs := res.Timestamp.Unix() - counter.Timestamp.Unix()

		if secs != 0 {
			rate = delta / float64(secs)
		}
	}

	// Store new counter value
	p.counterCache.Set(key, &CounterValue{
		Key:       key,
		Value:     newMetric,
		Timestamp: res.Timestamp,
	})

	return rate
}

func (p *Processor) getCounterKey(res *Resource, name string) string {
	key := res.Type
	if res.PodMetadata != nil {
		key += "-" + string(res.PodMetadata.Pod.UID)
	}
	key += "-" + res.Name + "-" + name

	return key
}

func (p *Processor) UptimeMetrics(startTime time.Time) Metrics {
	var uptime uint64
	if !startTime.IsZero() {
		uptime = uint64(time.Since(startTime).Nanoseconds() / time.Millisecond.Nanoseconds())
	}

	return Metrics{
		MeasureUptime: &Metric{Type: MetricTypeInt, IntValue: &uptime},
	}

}

func (p *Processor) CpuMetrics(s *stats.CPUStats, limit float64) Metrics {
	if s == nil {
		return nil
	}
	// Convert nanoCores to seconds
	var cpuUsage float64
	nanoCores := s.UsageNanoCores

	if nanoCores != nil {
		cpuUsage = float64(*nanoCores) / 1000000000
	}

	// if resource limits defined, get utilization
	var cpuUtilization float64
	if limit > 0 {
		cpuUtilization = (cpuUsage / limit) * 100
	}

	return Metrics{
		MeasureCpuUsage:       &Metric{Type: MetricTypeFloat, FloatValue: &cpuUsage},
		MeasureCpuUtilization: &Metric{Type: MetricTypeFloat, FloatValue: &cpuUtilization},
	}
}

func (p *Processor) MemMetrics(s *stats.MemoryStats, limit float64) Metrics {
	// MemoryStats are optional
	if s == nil {
		return Metrics{}
	}
	// if resource limits defined get utilization
	var utilization float64
	if limit > 0 {
		if s.UsageBytes != nil {
			usage := float64(*s.UsageBytes)
			utilization = (usage / limit) * 100
		}
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
		MeasureVolumeUsed:        &Metric{Type: MetricTypeInt, IntValue: s.UsedBytes},
		MeasureVolumeInodesTotal: &Metric{Type: MetricTypeInt, IntValue: s.Inodes},
		MeasureVolumeInodesFree:  &Metric{Type: MetricTypeInt, IntValue: s.InodesFree},
		MeasureVolumeInodesUsed:  &Metric{Type: MetricTypeInt, IntValue: s.InodesUsed},
	}
}
