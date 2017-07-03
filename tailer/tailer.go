package tailer

import (
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/handlers"
	"github.com/hpcloud/tail"
)

type Tailer struct {
	path    string
	done    chan bool
	handler handlers.LineHandler
}

func NewTailer(path string, handler handlers.LineHandler) *Tailer {
	t := &Tailer{
		path:    path,
		handler: handler,
		done:    make(chan bool),
	}
	return t
}

func (t *Tailer) Run() error {
	tailConf := tail.Config{
		ReOpen: true,
		Follow: true,
		Logger: tail.DiscardingLogger,
	}
	tailer, err := tail.TailFile(t.path, tailConf)
	if err != nil {
		logrus.WithField("filePath", t.path).Info("Error starting file tail")
		return err
	}
	logrus.WithField("filePath", t.path).Info("Tailing file")
	out := make(chan string)
	go func() {
	loop:
		for {
			select {
			case line, ok := <-tailer.Lines:
				if !ok {
					close(out)
					break loop
				}
				if line.Err != nil {
					continue
				}
				t.handler.Handle(line.Text)
			case <-t.done:
				close(out)
				break loop
			}
		}
		logrus.WithField("filePath", t.path).Info("Done tailing file")
	}()
	return nil
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
	checkInterval  time.Duration
	done           chan bool
}

func NewPathWatcher(pattern string, filter filterFunc, handlerFactory handlers.LineHandlerFactory) *PathWatcher {
	p := &PathWatcher{
		pattern:        pattern,
		filter:         filter,
		watched:        make(map[string]struct{}),
		handlerFactory: handlerFactory,
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
			tailer := NewTailer(file, handler)
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
