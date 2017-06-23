package tailer

import (
	"github.com/Sirupsen/logrus"
	"github.com/hpcloud/tail"
)

type LineHandler interface {
	Handle(lines <-chan string)
}

type Tailer struct {
	path    string
	done    chan bool
	handler LineHandler
}

func NewTailer(path string, handler LineHandler) *Tailer {
	return &Tailer{
		path:    path,
		handler: handler,
		done:    make(chan bool),
	}
}

func (t *Tailer) Run() error {
	lines, err := tailFile(t.path, t.done)
	if err != nil {
		return err
	}
	t.handler.Handle(lines)
	return nil
}

func tailFile(filePath string, done <-chan bool) (chan string, error) {
	tailConf := tail.Config{
		ReOpen: true,
		Follow: true,
		Logger: tail.DiscardingLogger,
	}
	tailer, err := tail.TailFile(filePath, tailConf)
	if err != nil {
		logrus.WithField("filePath", filePath).Info("Error starting file tail")
		return nil, err
	}
	logrus.WithField("filePath", filePath).Info("Tailing file")
	out := make(chan string)
	go func() {
		for {
			select {
			case line, ok := <-tailer.Lines:
				if !ok {
					close(out)
					return
				}
				if line.Err != nil {
					continue
				}
				out <- line.Text
			case <-done:
				close(out)
				return
			}
		}
	}()
	return out, nil
}
