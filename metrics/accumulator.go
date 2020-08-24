package metrics

import (
	"github.com/sirupsen/logrus"
	"time"

	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

type MetricGroup string

const (
	ContainerMetricGroup = MetricGroup("container")
	PodMetricGroup       = MetricGroup("pod")
	NodeMetricGroup      = MetricGroup("node")
	VolumeMetricGroup    = MetricGroup("volume")
)

var ValidMetricGroups = map[MetricGroup]bool{
	ContainerMetricGroup: true,
	PodMetricGroup:       true,
	NodeMetricGroup:      true,
	VolumeMetricGroup:    true,
}

type MetricDataAccumulator struct {
	Data                  []*ResourceMetrics
	mp                    *Processor
	metadata              *Metadata
	metricGroupsToCollect map[MetricGroup]bool
	time                  time.Time
	logger                *logrus.Logger
}

func (a *MetricDataAccumulator) nodeStats(nodeResource *Resource, s stats.NodeStats) {
	if !a.metricGroupsToCollect[NodeMetricGroup] {
		return
	}

	a.accumulate(
		nodeResource,
		a.mp.UptimeMetrics(s.StartTime.Time),
		a.mp.CpuMetrics(s.CPU),
		a.mp.FsMetrics(s.Fs),
		a.mp.MemMetrics(s.Memory),
		a.mp.NetworkMetrics(s.Network),
	)
}

func (a *MetricDataAccumulator) podStats(podResource *Resource, s stats.PodStats) {
	if !a.metricGroupsToCollect[PodMetricGroup] {
		return
	}

	a.accumulate(
		podResource,
		a.mp.UptimeMetrics(s.StartTime.Time),
		a.mp.CpuMetrics(s.CPU),
		a.mp.FsMetrics(s.EphemeralStorage),
		a.mp.MemMetrics(s.Memory),
		a.mp.NetworkMetrics(s.Network),
	)
}

func (a *MetricDataAccumulator) containerStats(podResource *Resource, s stats.ContainerStats) {
	if !a.metricGroupsToCollect[ContainerMetricGroup] {
		return
	}

	resource, err := getContainerResource(podResource, s, a.metadata)
	if err != nil {
		a.logger.WithFields(logrus.Fields{
			"pod":       podResource.Labels[LabelPodName],
			"container": podResource.Labels[LabelContainerName],
		}).Warn("failed to fetch container metrics")
		return
	}

	a.accumulate(
		resource,
		a.mp.UptimeMetrics(s.StartTime.Time),
		a.mp.CpuMetrics(s.CPU),
		a.mp.MemMetrics(s.Memory),
		a.mp.FsMetrics(s.Rootfs),
	)
}

func (a *MetricDataAccumulator) volumeStats(podResource *Resource, s stats.VolumeStats) {
	if !a.metricGroupsToCollect[VolumeMetricGroup] {
		return
	}

	volume := getVolumeResource(podResource, s, a.metadata)

	a.accumulate(
		volume,
		a.mp.VolumeMetrics(s),
	)
}

func (a *MetricDataAccumulator) accumulate(r *Resource, m ...Metrics) {
	retMetrics := make(Metrics)

	for _, metrics := range m {
		for k, v := range metrics {
			retMetrics[k] = v
		}
	}
	a.Data = append(a.Data, &ResourceMetrics{
		Resource: r,
		Metrics:  retMetrics,
	})
}
