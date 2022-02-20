// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package service

import (
	"context"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
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
	dataset           string
	interval          time.Duration
	k8sClusterName    string
	restClient        kubelet.RestClient
	omitLabels        []metrics.OmitLabel
	additionalFields  map[string]interface{}
	includeNodeLabels bool

	metricGroupsToCollect map[metrics.MetricGroup]bool
	transmitter           *transmission.HoneycombTransmitter
	apiClient             corev1.NodesGetter
}

func newRunnable(rc kubelet.RestClient, opt Options, client *corev1.CoreV1Client) *runnable {
	return &runnable{
		dataset:               opt.Dataset,
		interval:              opt.Interval,
		k8sClusterName:        opt.ClusterName,
		restClient:            rc,
		omitLabels:            opt.OmitLabels,
		additionalFields:      opt.AdditionalFields,
		includeNodeLabels:     opt.IncludeNodeLabels,
		metricGroupsToCollect: opt.MetricGroupsToCollect,
		transmitter:           &transmission.HoneycombTransmitter{},
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

		logrus.WithFields(logrus.Fields{
			"resourceType": rm.Resource.Type,
			"resourceName": rm.Resource.Name,
		}).Trace("Creating event for resource")

		// create event from Resource meta
		ev := r.createEventFromResource(rm.Resource)

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
			ev.Data[pre+k] = val
		}

		logrus.WithFields(logrus.Fields{
			"resourceType": rm.Resource.Type,
			"resourceName": rm.Resource.Name,
			"metricCount":  len(rm.Metrics),
		}).Trace("Event Data: ", ev.Data)

		// send resource event to Honeycomb
		r.transmitter.Send(ev)
	}

	return nil
}

func (r *runnable) createEventFromResource(res *metrics.Resource) *event.Event {

	// add basic attributes to each resource event
	data := map[string]interface{}{
		metrics.PrefixMetrics + metrics.MetricSourceName: res.Name,
		metrics.PrefixMetrics + metrics.MetricSourceType: res.Type,
		metrics.KubernetesResourceType:                   res.Type,
		metrics.PrefixCluster + "name":                   r.k8sClusterName,
	}

	// add all additional fields
	for k, v := range r.additionalFields {
		data[k] = v
	}

	// add all labels to event
	for k, v := range res.Labels {
		data[k] = v
	}

	// add status details to event
	if res.Status != nil {
		for k, v := range res.Status {
			data[k] = v
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
					data[metrics.StatusRestart] = true
				}
				data[metrics.StatusRestartDelta] = delta
			}
		}
	}

	ev := &event.Event{
		Data:      data,
		Dataset:   r.dataset,
		Timestamp: res.Timestamp,
	}

	return ev
}
