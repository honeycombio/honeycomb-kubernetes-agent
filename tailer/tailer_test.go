package tailer

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockLineHandler struct {
	lines []string
}

func (m *mockLineHandler) Handle(line string) {
	m.lines = append(m.lines, line)
}

func TestTailing(t *testing.T) {
	logFile, err := ioutil.TempFile("/tmp", "honeycomb-log-test")
	assert.NoError(t, err)

	stateFile, err := ioutil.TempFile("/tmp", "honeycomb-log-test-statefile")
	assert.NoError(t, err)

	stateRecorder, err := NewStateRecorder(stateFile.Name())
	assert.NoError(t, err)

	handler := &mockLineHandler{}

	logFile.Write([]byte("line1\n"))
	logFile.Sync()
	tailer := NewTailer(logFile.Name(), handler, stateRecorder)
	tailer.Run()

	logFile.Write([]byte("line2\n"))
	logFile.Sync()
	time.Sleep(time.Second)
	tailer.Stop()

	assert.Equal(t, len(handler.lines), 2)
	assert.Equal(t, handler.lines[0], "line1")
	assert.Equal(t, handler.lines[1], "line2")

	tailer = NewTailer(logFile.Name(), handler, stateRecorder)
	tailer.Run()
	tailer.Stop()
	assert.Equal(t, len(handler.lines), 2)
	assert.Equal(t, handler.lines[0], "line1")
	assert.Equal(t, handler.lines[1], "line2")

	tailer.Run()
	logFile.Write([]byte("line3\n"))
	logFile.Sync()
	logFile.Close()
	os.Remove(logFile.Name())

	time.Sleep(time.Second)

	_, err = stateRecorder.Get(logFile.Name())
	assert.Error(t, err)
}
