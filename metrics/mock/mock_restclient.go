package metrics_mock

import "io/ioutil"

type MockRestClient struct {
}

func (f MockRestClient) StatsSummary() ([]byte, error) {
	return ioutil.ReadFile("../testdata/stats-summary.json")
}

func (f MockRestClient) Pods() ([]byte, error) {
	return ioutil.ReadFile("../testdata/pods.json")
}
