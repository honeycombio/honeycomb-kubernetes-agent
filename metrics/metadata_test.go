package metrics

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/kubelet"
	"github.com/honeycombio/honeycomb-kubernetes-agent/metrics/mock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func mockMetaData(omitLabels []OmitLabel) *Metadata {
	rc := &metrics_mock.MockRestClient{}
	metadataProvider := kubelet.NewMetadataProvider(rc)
	pods, _ := metadataProvider.Pods()
	return NewMetadata(pods, omitLabels, logrus.StandardLogger())
}

func TestGetLabels(t *testing.T) {

	md := mockMetaData([]OmitLabel{
		"controller-revision-hash",
		"pod-template-generation",
	})
	pmd, _ := md.GetPodMetadataByUid("5c69ffd4-73d1-47df-87d7-08ae860e75ae")

	labels := pmd.GetLabels()

	assert.Equal(t, 3, len(labels))

	assert.Equal(t, "honeycomb", labels["app.kubernetes.io/name"])
}

func TestGetCpuLimit(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	limit := pmd.GetCpuLimit()

	assert.Equal(t, 0.1, limit)
}

func TestGetCpuLimitForContainer(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	limit := pmd.GetCpuLimitForContainer("speaker")

	assert.Equal(t, 0.1, limit)
}

func TestGetMemoryLimit(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	limit := pmd.GetMemoryLimit()

	assert.Equal(t, 100*math.Pow(2, 20), limit)
}

func TestGetMemoryLimitForContainer(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	limit := pmd.GetMemoryLimitForContainer("speaker")

	assert.Equal(t, 100*math.Pow(2, 20), limit)
}

func TestGetStatus(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	status := pmd.GetStatus()

	assert.Equal(t, "Running", status[StatusPodPhase])
}

func TestGetStatusForContainer(t *testing.T) {
	md := mockMetaData(nil)
	pmd, _ := md.GetPodMetadataByUid("5997ad9b-1d2a-43cf-ab57-a98d8796dc34")

	status := pmd.GetStatusForContainer("speaker")

	assert.Equal(t, "54", status[StatusContainerRestarts])
	assert.Equal(t, "true", status[StatusContainerReady])
	assert.Equal(t, "running", status[StatusContainerState])
}
