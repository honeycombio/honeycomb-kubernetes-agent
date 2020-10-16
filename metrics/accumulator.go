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
	a.logger.WithFields(logrus.Fields{
		"name": nodeResource.Name,
	}).Debug("nodeStats")

	if !a.metricGroupsToCollect[NodeMetricGroup] {
		return
	}

	a.accumulate(
		nodeResource,
		a.mp.UptimeMetrics(s.StartTime.Time),
		a.mp.CpuMetrics(s.CPU, 0),
		a.mp.FsMetrics(s.Fs),
		a.mp.MemMetrics(s.Memory, float64(*s.Memory.AvailableBytes)),
		a.mp.NetworkMetrics(s.Network),
	)
}

func (a *MetricDataAccumulator) podStats(podResource *Resource, s stats.PodStats) {
	a.logger.WithFields(logrus.Fields{
		"name": podResource.Name,
	}).Debug("podStats")

	if !a.metricGroupsToCollect[PodMetricGroup] {
		return
	}

	a.accumulate(
		podResource,
		a.mp.UptimeMetrics(s.StartTime.Time),
		a.mp.CpuMetrics(s.CPU, podResource.PodMetadata.GetCpuLimit()),
		a.mp.FsMetrics(s.EphemeralStorage),
		a.mp.MemMetrics(s.Memory, podResource.PodMetadata.GetMemoryLimit()),
		a.mp.NetworkMetrics(s.Network),
	)
}

func (a *MetricDataAccumulator) containerStats(podResource *Resource, s stats.ContainerStats) {
	a.logger.WithFields(logrus.Fields{
		"podName": podResource.Name,
		"name":    s.Name,
	}).Debug("containerStats")

	if !a.metricGroupsToCollect[ContainerMetricGroup] {
		return
	}

	if s.CPU == nil {
		return
	}

	resource, err := getContainerResource(podResource, s)

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
		a.mp.CpuMetrics(s.CPU, resource.PodMetadata.GetCpuLimitForContainer(s.Name)),
		a.mp.MemMetrics(s.Memory, podResource.PodMetadata.GetMemoryLimitForContainer(s.Name)),
		a.mp.FsMetrics(s.Rootfs),
	)
}

func (a *MetricDataAccumulator) volumeStats(podResource *Resource, s stats.VolumeStats) {
	a.logger.WithFields(logrus.Fields{
		"podName": podResource.Name,
		"name":    s.Name,
	}).Debug("volumeStats")

	if !a.metricGroupsToCollect[VolumeMetricGroup] {
		return
	}

	volume := getVolumeResource(podResource, s)

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
