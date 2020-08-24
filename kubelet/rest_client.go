// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

// RestClient is swappable for testing.
type RestClient interface {
	StatsSummary() ([]byte, error)
	Pods() ([]byte, error)
}

// RestClient is a thin wrapper around a kubelet client, encapsulating endpoints
// and their corresponding http methods. The endpoints /stats/container /spec/
// are excluded because they require cadvisor. The /metrics endpoint is excluded
// because it returns Prometheus data.
type HTTPRestClient struct {
	client Client
}

func NewRestClient(client Client) *HTTPRestClient {
	return &HTTPRestClient{client: client}
}

func (c *HTTPRestClient) StatsSummary() ([]byte, error) {
	return c.client.Get("/stats/summary")
}

func (c *HTTPRestClient) Pods() ([]byte, error) {
	return c.client.Get("/pods")
}
