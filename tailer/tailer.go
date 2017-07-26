package tailer

import (
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/hpcloud/tail"
)

type Tailer struct {
	path          string
	done          chan bool
	handler       handlers.LineHandler
	stateRecorder StateRecorder
}

func NewTailer(path string, handler handlers.LineHandler, stateRecorder StateRecorder) *Tailer {
	t := &Tailer{
		path:          path,
		handler:       handler,
		stateRecorder: stateRecorder,
		done:          make(chan bool),
	}
	return t
}

func (t *Tailer) Run() error {
	seekInfo := &tail.SeekInfo{}
	if t.stateRecorder != nil {
		if offset, err := t.stateRecorder.Get(t.path); err == nil {
			seekInfo.Offset = offset
		}
	}
	tailConf := tail.Config{
		ReOpen: false,
		Follow: true,
		// TODO: inotify doesn't detect file deletions, fix this
		Poll:     true,
		Logger:   tail.DiscardingLogger,
		Location: seekInfo,
	}
	tailer, err := tail.TailFile(t.path, tailConf)
	if err != nil {
		logrus.WithField("filePath", t.path).Info("Error starting file tail")
		return err
	}
	logrus.WithField("path", t.path).WithField("offset", tailConf.Location.Offset).
		Info("Tailing file")
	ticker := time.NewTicker(time.Second)
	go func() {
	loop:
		for {
			select {
			case line, ok := <-tailer.Lines:
				if !ok {
					if t.stateRecorder != nil {
						t.stateRecorder.Delete(t.path)
					}
					break loop
				}
				if line.Err != nil {
					continue
				}
				t.handler.Handle(line.Text)
			case <-t.done:
				t.updateState(tailer.Tell)
				ticker.Stop()
				break loop
			case <-ticker.C:
				t.updateState(tailer.Tell)
			}
		}
		logrus.WithField("filePath", t.path).Info("Done tailing file")
	}()
	return nil
}

func (t *Tailer) updateState(teller func() (int64, error)) {
	offset, err := teller()
	if err == nil && t.stateRecorder != nil {
		t.stateRecorder.Record(t.path, offset)
	} else {
		// TODO
	}

}

func (t *Tailer) Stop() {
	t.done <- true
}

type filterFunc func(string) bool

type PathWatcher struct {
	pattern        string
	filter         filterFunc
	watched        map[string]struct{}
	handlerFactory handlers.LineHandlerFactory
	stateRecorder  StateRecorder
	checkInterval  time.Duration
	done           chan bool
}

func NewPathWatcher(
	pattern string,
	filter filterFunc,
	handlerFactory handlers.LineHandlerFactory,
	stateRecorder StateRecorder,
) *PathWatcher {
	p := &PathWatcher{
		pattern:        pattern,
		filter:         filter,
		watched:        make(map[string]struct{}),
		handlerFactory: handlerFactory,
		stateRecorder:  stateRecorder,
		checkInterval:  time.Second, // TODO make configurable
		done:           make(chan bool),
	}

	return p
}

func (p *PathWatcher) Run() {
	ticker := time.NewTicker(p.checkInterval)
	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.check()
		}
	}
}

func (p *PathWatcher) Stop() {
	p.done <- true
}

func (p *PathWatcher) check() {
	files, err := filepath.Glob(p.pattern)
	if err != nil {
		logrus.WithError(err).Error("Error globbing files")
	}
	current := make(map[string]struct{}, len(p.watched))
	for _, file := range files {
		_, ok := p.watched[file]
		if !ok {
			if p.filter != nil && !p.filter(file) {
				continue
			}
			p.watched[file] = struct{}{}
			handler := p.handlerFactory.New(file)
			tailer := NewTailer(file, handler, p.stateRecorder)
			go tailer.Run()
		}
		current[file] = struct{}{}
	}
	for file := range p.watched {
		_, ok := current[file]
		if !ok {
			delete(p.watched, file)
		}
	}
}
