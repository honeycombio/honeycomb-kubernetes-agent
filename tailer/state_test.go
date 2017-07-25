package tailer

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateRecorder(t *testing.T) {
	stateFileHandle, err := ioutil.TempFile("/tmp", "honeycomb-agent-state")
	assert.NoError(t, err)
	stateFile := stateFileHandle.Name()
	sr, err := NewStateRecorder(stateFile)
	assert.NoError(t, err)
	err = sr.Record("/var/log/wherever", 22)
	assert.NoError(t, err)
	offset, err := sr.Get("/var/log/wherever")
	assert.NoError(t, err)
	assert.Equal(t, offset, int64(22))
}
