package tailer

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hpcloud/tail"
)

func TailFile(filePath string, out chan string) error {
	tailConf := tail.Config{
		ReOpen: true,
		Follow: true,
		Logger: tail.DiscardingLogger,
	}
	tailer, err := tail.TailFile(filePath, tailConf)
	if err != nil {
		logrus.WithField("filePath", filePath).Info("Error starting file tail")
		return err
	}
	logrus.WithField("filePath", filePath).Info("Tailing file")
	for {
		select {
		case line, ok := <-tailer.Lines:
			if !ok {
				return nil
			}
			if line.Err != nil {
				continue
			}
			out <- strings.TrimSpace(line.Text)
		}
	}
}
