package metrics

import (
	"time"

	"github.com/sirupsen/logrus"

	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
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
	Data                  map[string]*ResourceMetrics
	mp                    *Processor
	metadata              *Metadata
	metricGroupsToCollect map[MetricGroup]bool
	time                  time.Time
}

func (a *MetricDataAccumulator) nodeStats(nodeResource *Resource, s stats.NodeStats) {
	logrus.WithFields(logrus.Fields{
		"name": nodeResource.Name,
	}).Trace("nodeStats")

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
	logrus.WithFields(logrus.Fields{
		"name": podResource.Name,
	}).Trace("podStats")

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
	logrus.WithFields(logrus.Fields{
		"podName": podResource.Name,
		"name":    s.Name,
	}).Trace("containerStats")

	if !a.metricGroupsToCollect[ContainerMetricGroup] {
		return
	}

	if s.CPU == nil {
		return
	}

	resource, err := getContainerResource(podResource, s)

	if err != nil {
		logrus.WithFields(logrus.Fields{
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
	logrus.WithFields(logrus.Fields{
		"podName": podResource.Name,
		"name":    s.Name,
	}).Trace("volumeStats")

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
	podName := r.Name

	// check if pod name already exists
	var resourceMetrics *ResourceMetrics
	if rm, ok := a.Data[podName]; ok {
		// use existing entry from map
		resourceMetrics = rm
	} else {
		// create new entry and add to map
		resourceMetrics = &ResourceMetrics{
			Resource: r,
			Metrics:  make(Metrics),
		}
		a.Data[podName] = resourceMetrics
	}

	// add metrics to resourceMetrics
	for _, metric := range m {
		for k, v := range metric {
			resourceMetrics.Metrics[k] = v
		}
	}
}
