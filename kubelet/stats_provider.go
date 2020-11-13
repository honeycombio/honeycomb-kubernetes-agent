// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

import (
	"encoding/json"

	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

// StatsProvider wraps a RestClient, returning an unmarshalled
// stats.Summary struct from the kubelet API.
type StatsProvider struct {
	rc RestClient
}

func NewStatsProvider(rc RestClient) *StatsProvider {
	return &StatsProvider{rc: rc}
}

// StatsSummary calls the /stats/summary kubelet endpoint and unmarshalls the
// results into a stats.Summary struct.
func (p *StatsProvider) StatsSummary() (*stats.Summary, error) {
	summary, err := p.rc.StatsSummary()
	if err != nil {
		return nil, err
	}
	var out stats.Summary
	err = json.Unmarshal(summary, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
