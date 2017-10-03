package tailer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/stretchr/testify/assert"
)

type mockLineHandler struct {
	lines []string
}

func (m *mockLineHandler) Handle(line string) {
	m.lines = append(m.lines, line)
}

type mockLineHandlerFactory struct {
	handlers map[string]*mockLineHandler
}

func newMockLineHandlerFactory() *mockLineHandlerFactory {
	return &mockLineHandlerFactory{
		handlers: make(map[string]*mockLineHandler),
	}
}

func (mf *mockLineHandlerFactory) New(path string) handlers.LineHandler {
	h := &mockLineHandler{}
	mf.handlers[path] = h
	return h
}

type logger struct {
	counter int
	path    string
	file    *os.File
}

func (l *logger) Write() error {
	if l.file == nil {
		var err error
		l.file, err = os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
		if err != nil {
			return err
		}
	}
	_, err := l.file.Write([]byte(fmt.Sprintf("line %d\n", l.counter)))
	if err != nil {
		return err
	}
	l.counter++
	return nil
}

func (l *logger) Rotate() error {
	if err := l.file.Close(); err != nil {
		return err
	}
	if err := os.Remove(l.path); err != nil {
		return err
	}
	l.file = nil
	return nil
}

func TestTailRestarting(t *testing.T) {
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
}

func TestPathWatching(t *testing.T) {
	dir := "/tmp/honeycomb-log-test/"
	stateFile, err := ioutil.TempFile("/tmp", "honeycomb-log-test-statefile")
	assert.NoError(t, err)
	stateRecorder, err := NewStateRecorder(stateFile.Name())
	assert.NoError(t, err)

	handlerFactory := newMockLineHandlerFactory()

	watcher := NewPathWatcher(
		fmt.Sprintf("%s/*", dir),
		nil,
		handlerFactory,
		stateRecorder,
	)

	watcher.checkInterval = time.Millisecond
	go watcher.Run()

	err = os.MkdirAll(dir, os.FileMode(0755))
	assert.NoError(t, err)

	loggers := make([]*logger, 2)
	for i := 0; i < 2; i++ {
		fileName := path.Join(dir, fmt.Sprintf("%d.log", i))
		loggers[i] = &logger{path: fileName}
		err := loggers[i].Write()
		assert.NoError(t, err)
	}

	time.Sleep(time.Second)

	for _, l := range loggers {
		l.Rotate()
		l.Write()
	}
	time.Sleep(time.Second)

	for _, l := range loggers {
		l.Rotate()
	}
	time.Sleep(time.Second)

	assert.Equal(t, len(handlerFactory.handlers), 2)
	for _, h := range handlerFactory.handlers {
		assert.Equal(t, len(h.lines), 2)
	}
	assert.Equal(t, len(watcher.tailers), 0)
}
