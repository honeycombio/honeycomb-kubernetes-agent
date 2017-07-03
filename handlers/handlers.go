package handlers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
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
	unwrapper     unwrappers.Unwrapper
	parserFactory parsers.ParserFactory
	processors    []processors.Processor
}

func NewLineHandlerFactoryFromConfig(
	config *config.WatcherConfig,
	unwrapper unwrappers.Unwrapper,
	extraProcessors ...processors.Processor,
) (*LineHandlerFactoryImpl, error) {
	ret := &LineHandlerFactoryImpl{
		config:    config,
		unwrapper: unwrapper,
	}
	parserFactory, err := parsers.NewParserFactory(config.Parser)
	if err != nil {
		return nil, fmt.Errorf("Error setting up parser: %v", err)
	}
	ret.parserFactory = parserFactory

	for _, processorConfig := range config.Processors {
		processor, err := processors.NewProcessorFromConfig(processorConfig)
		if err != nil {
			return nil, fmt.Errorf("Error setting up processor: %v", err)
		}
		ret.processors = append(ret.processors, processor)
	}
	ret.processors = append(ret.processors, extraProcessors...)

	return ret, nil
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
	unwrapper  unwrappers.Unwrapper
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
