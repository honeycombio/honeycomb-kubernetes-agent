package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	libhoney "github.com/honeycombio/libhoney-go"
)

type LineHandler interface {
	Handle(string)
}

type LineHandlerFactory interface {
	New(path string) LineHandler
}

type LineHandlerFactoryImpl struct {
	config        *config.WatcherConfig
	parserFactory parsers.ParserFactory
	processors    []processors.Processor
	unwrapper     Unwrapper
}

func NewLineHandlerFactory(
	config *config.WatcherConfig,
	parserFactory parsers.ParserFactory,
	unwrapper Unwrapper,
	processors ...processors.Processor) *LineHandlerFactoryImpl {
	return &LineHandlerFactoryImpl{
		config:        config,
		parserFactory: parserFactory,
		unwrapper:     unwrapper,
		processors:    processors,
	}
}

func (hf *LineHandlerFactoryImpl) New(path string) LineHandler {
	handler := &LineHandlerImpl{
		config:     hf.config,
		parser:     hf.parserFactory.New(),
		processors: hf.processors,
		unwrapper:  hf.unwrapper,
	}
	handler.builder = libhoney.NewBuilder()
	handler.builder.Dataset = handler.config.Dataset
	return handler
}

type LineHandlerImpl struct {
	config     *config.WatcherConfig
	unwrapper  Unwrapper
	parser     parsers.Parser
	processors []processors.Processor
	builder    *libhoney.Builder
}

func (h *LineHandlerImpl) Handle(rawLine string) {
	parsed, err := h.unwrapper.Unwrap(rawLine, h.parser)
	if err != nil {
		logrus.WithError(err).Debug("Failed to parse line")
		return
	}
	for _, p := range h.processors {
		p.Process(parsed)
	}
	logrus.WithField("parsed", parsed).Debug("Sending line")
	h.builder.SendNow(parsed)
}

type Unwrapper interface {
	Unwrap(string, parsers.Parser) (map[string]interface{}, error)
}

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

type RawLogUnwrapper struct{}

func (u *RawLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (map[string]interface{}, error) {
	return parser.Parse(rawLine)
}
