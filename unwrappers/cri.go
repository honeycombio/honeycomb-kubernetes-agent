package unwrappers

import (
	"fmt"
	"strings"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/sirupsen/logrus"
)

type criLogLine struct {
	Time   string
	Stream string
	Tags   string
	Log    string
}

type CriLogUnwrapper struct{}

// 2020-04-04T03:20:26.7063258Z stdout F {rest of message follows}

func (u *CriLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (*event.Event, error) {
	line := &criLogLine{}
	parts := strings.SplitN(rawLine, " ", 4)
	if len(parts) != 4 {
		logrus.Info("Error parsing CRI line")
		return nil, fmt.Errorf("Error parsing log line as CRI log: '%s'", line)
	}
	line.Time = parts[0]
	line.Stream = parts[1]
	line.Tags = parts[2]
	line.Log = parts[3]
	line.Log = strings.TrimRight(line.Log, "\n")

	data, err := parser.Parse(line.Log)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	ts, err := time.Parse(time.RFC3339Nano, line.Time)
	if err != nil {
		logrus.WithError(err).Info("Error parsing CRI timestamp")
	}

	return &event.Event{
		Data:       data,
		Timestamp:  ts,
		RawMessage: line.Log,
	}, nil
}
