// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package service

import (
	"context"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/honeycombio/honeycomb-kubernetes-agent/interval"
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
)

var _ interval.Runnable = (*runnable)(nil)

type runnable struct {
	statsProvider     *kubelet.StatsProvider
	metadataProvider  *kubelet.MetadataProvider
	metricsProvider   *metrics.Processor
	interval          time.Duration
	k8sClusterName    string
	restClient        kubelet.RestClient
	builder           *libhoney.Builder
	omitLabels        []metrics.OmitLabel
	includeNodeLabels bool

	metricGroupsToCollect map[metrics.MetricGroup]bool
	apiClient             corev1.NodesGetter
}

func newRunnable(rc kubelet.RestClient, builder *libhoney.Builder, opt Options, client *corev1.CoreV1Client) *runnable {
	return &runnable{
		interval:              opt.Interval,
		k8sClusterName:        opt.ClusterName,
		restClient:            rc,
		builder:               builder,
		omitLabels:            opt.OmitLabels,
		includeNodeLabels:     opt.IncludeNodeLabels,
		metricGroupsToCollect: opt.MetricGroupsToCollect,
		apiClient:             client,
	}
}

// Sets up the kubelet connection at startup time.
func (r *runnable) Setup() error {
	logrus.Info("Creating Metrics Service Providers...")

	r.statsProvider = kubelet.NewStatsProvider(r.restClient)
	r.metadataProvider = kubelet.NewMetadataProvider(r.restClient)
	r.metricsProvider = metrics.NewMetricsProcessor(r.interval)

	logrus.Debug("Metrics Service Providers created")
	return nil
}

func (r *runnable) Run() error {

	// get all stats from stats provider
	summary, err := r.statsProvider.StatsSummary()
	if err != nil {
		logrus.WithError(err).Error("Could not retrieve stats")
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"podCount": len(summary.Pods),
	}).Debug("Retrieved Stats")

	// fetch metadata for all pods/containers
	var podsMetadata *v1.PodList
	podsMetadata, err = r.metadataProvider.Pods()
	if err != nil {
		logrus.WithError(err).Error("Could not retrieve pod metadata")
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"podCount": len(podsMetadata.Items),
	}).Debug("Retrieved Pod Metadata")

	var nodesMetadata *v1.NodeList
	// fetch metadata for all nodes
	if r.includeNodeLabels {
		nodesMetadata, err = r.apiClient.Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.WithError(err).Error("Could not retrieve node metadata")
			return nil
		}
		logrus.WithFields(logrus.Fields{
			"nodeCount": len(nodesMetadata.Items),
		}).Debug("Retrieved Node Metadata")
	}

	// Get resource metrics
	metadata := metrics.NewMetadata(podsMetadata, nodesMetadata, r.omitLabels, r.includeNodeLabels)
	resourceMetrics := r.metricsProvider.GenerateMetricsData(summary, metadata, r.metricGroupsToCollect)
	logrus.WithFields(logrus.Fields{
		"resourceCount": len(resourceMetrics),
	}).Debug("Processing Metrics Data...")

	// iterate over resource metrics data
	for _, rm := range resourceMetrics {

		// create event from Resource meta
		ev, err := r.createEventFromResource(rm.Resource)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"resourceType": rm.Resource.Type,
				"resourceName": rm.Resource.Name,
			}).WithError(err).Error("Could not create event for resource")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"resourceType": rm.Resource.Type,
			"resourceName": rm.Resource.Name,
		}).Trace("Creating event for resource")

		pre := metrics.PrefixMetrics

		// loop through all metrics to add them to resource event
		for k, v := range rm.Metrics {
			logrus.WithFields(logrus.Fields{
				"resourceType": rm.Resource.Type,
				"resourceName": rm.Resource.Name,
				"metricName":   k,
			}).Trace("Metric to event field")

			// check if metric is a counter
			var val float64
			if v.IsCounter {
				val = r.metricsProvider.GetCounterRate(rm.Resource, k, v)
			} else {
				val = v.GetValue()
			}

			// add metric to resource event
			ev.AddField(pre+k, val)
		}

		logrus.WithFields(logrus.Fields{
			"resourceType": rm.Resource.Type,
			"resourceName": rm.Resource.Name,
			"metricCount":  len(rm.Metrics),
		}).Debug("Event Data: ", ev.Fields())

		// send resource event to Honeycomb
		if err = ev.Send(); err != nil {
			logrus.WithFields(logrus.Fields{
				"resourceType": rm.Resource.Type,
				"resourceName": rm.Resource.Name,
				"metricCount":  len(rm.Metrics),
			}).WithError(err).Error("Error sending event to honeycomb")
		}
	}

	return nil
}

func (r *runnable) createEventFromResource(res *metrics.Resource) (*libhoney.Event, error) {
	ev := r.builder.NewEvent()

	// add basic attributes to each resource event
	_ = ev.Add(map[string]string{
		metrics.PrefixMetrics + metrics.MetricSourceName: res.Name,
		metrics.PrefixMetrics + metrics.MetricSourceType: res.Type,
		metrics.KubernetesResourceType:                   res.Type,
		metrics.PrefixCluster + "name":                   r.k8sClusterName,
	})
	ev.Timestamp = res.Timestamp

	// add all labels to event
	if err := ev.Add(res.Labels); err != nil {
		return nil, err
	}

	// add status attributes as fields in event
	if res.Status != nil {
		if err := ev.Add(res.Status); err != nil {
			return nil, err
		}

		// Check if resource restarted
		if rc, ok := res.Status[metrics.StatusRestartCount]; ok {
			// get current restart count
			if rci, ok := rc.(int32); ok {
				restarts := uint64(rci)
				m := &metrics.Metric{Type: metrics.MetricTypeInt, IsCounter: true, IntValue: &restarts}

				// use metricsProvider counter cache to check if restarts increased
				delta := r.metricsProvider.GetCounterDelta(res, "restarts", m)
				if delta > 0 {
					// resource had a restart since last collection
					ev.AddField(metrics.StatusRestart, true)
				}
				ev.AddField(metrics.StatusRestartDelta, delta)
			}
		}
	}

	return ev, nil
}
