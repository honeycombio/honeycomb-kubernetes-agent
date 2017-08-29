package unwrappers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
)

type dockerJSONLogLine struct {
	Log    string
	Stream string
	Time   string
}

type DockerJSONLogUnwrapper struct{}

func (u *DockerJSONLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (*event.Event, error) {
	line := &dockerJSONLogLine{}
	err := json.Unmarshal([]byte(rawLine), line)
	if err != nil {
		logrus.WithError(err).Info("Error parsing docker JSON line")
		return nil, fmt.Errorf("Error parsing log line as Docker json-file log: %v", err)
	}
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
		logrus.WithError(err).Info("Error parsing docker JSON timestamp")
	}

	return &event.Event{
		Data:      data,
		Timestamp: ts,
	}, nil
}
