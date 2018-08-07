package podtailer

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
)

func TestDetermineFilterFuncLegacyLogs(t *testing.T) {
	pod := &v1.Pod{}
	pod.Name = "foo"
	pod.Namespace = "default"
	expected := func(fileName string) bool {
		re := fmt.Sprintf(
			"^/var/log/containers/%s_%s_%s-.+\\.log",
			"foo",
			"default",
			"wibble",
		)
		ok, _ := regexp.Match(re, []byte(fileName))
		return ok
	}

	f := determineFilterFunc(pod, "wibble", true)
	// functions should have same output
	assert.Equal(
		t,
		expected("/var/log/containers/foo_default_wibble-abcdefg.log"),
		f("/var/log/containers/foo_default_wibble-abcdefg.log"),
	)

	assert.Equal(
		t,
		expected("/var/log/containers/foo_default_wobble-abcdefg.log"),
		f("/var/log/containers/foo_default_wobble-abcdefg.log"),
	)
}

func TestDetermineFilterFunc(t *testing.T) {
	pod := &v1.Pod{}
	pod.UID = "123456"

	expected := func(fileName string) bool {
		re := fmt.Sprintf(
			"^/var/log/pods/%s/%s_[0-9]*\\.log",
			"123456",
			"wibble",
		)
		ok, _ := regexp.Match(re, []byte(fileName))
		return ok
	}

	f := determineFilterFunc(pod, "wibble", false)
	assert.Equal(
		t,
		expected("/var/log/pods/123456/wibble_0.log"),
		f("/var/log/pods/123456/wibble_0.log"),
	)

	assert.Equal(
		t,
		expected("/var/log/pods/123456/wobble_0.log"),
		f("/var/log/pods/123456/wobble_0.log"),
	)
}
