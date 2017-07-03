package unwrappers

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
)

type dockerJSONLogLine struct {
	Log    string
	Stream string
	Time   string
}

type DockerJSONLogUnwrapper struct{}

func (u *DockerJSONLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (map[string]interface{}, error) {
	line := &dockerJSONLogLine{}
	err := json.Unmarshal([]byte(rawLine), line)
	if err != nil {
		logrus.WithError(err).Info("Error parsing JSON line")
		return nil, fmt.Errorf("Error parsing log line as Docker json-file log: %v", err)
	}

	return parser.Parse(line.Log)
}
