package handlers

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	libhoney "github.com/honeycombio/libhoney-go"
)

type LineHandler interface {
	Init()
	Handle(string)
}

type HandlerFactory interface {
	New(string) LineHandler
	AddProcessor(processors.Processor)
}

type DockerJSONLogHandlerFactory struct {
	Config        *config.WatcherConfig
	ParserFactory parsers.ParserFactory
	Processors    []processors.Processor
}

func (hf *DockerJSONLogHandlerFactory) AddProcessor(p processors.Processor) {
	hf.Processors = append(hf.Processors, p)
}

func (hf *DockerJSONLogHandlerFactory) New(path string) LineHandler {
	handler := &DockerJSONLogHandler{
		Config:     hf.Config,
		Parser:     hf.ParserFactory.New(),
		Processors: hf.Processors,
	}
	handler.Init()
	return handler
}

type DockerJSONLogHandler struct {
	Config     *config.WatcherConfig
	Parser     parsers.Parser
	Processors []processors.Processor
	builder    *libhoney.Builder
}

type dockerJSONLogLine struct {
	Log    string
	Stream string
	Time   string
}

func (h *DockerJSONLogHandler) Init() {
	h.builder = libhoney.NewBuilder()
	h.builder.Dataset = h.Config.Dataset
}

func (h *DockerJSONLogHandler) Handle(rawLine string) {
	// multiline parsing should be done with stateful parsers for now
	line := &dockerJSONLogLine{}
	err := json.Unmarshal([]byte(rawLine), line)
	if err != nil {
		logrus.WithError(err).Info("Error parsing JSON line")
		return
	}

	parsed, err := h.Parser.Parse(line.Log)
	if err != nil {
		logrus.WithError(err).Debug("Failed to parse line")
		return
	}
	for _, p := range h.Processors {
		p.Process(parsed)
	}
	logrus.WithField("parsed", parsed).Debug("Sending line")
	h.builder.SendNow(parsed)
}
