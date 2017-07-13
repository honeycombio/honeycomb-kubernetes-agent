// Package handlers drives the actual processing of log lines. For each set of
// files that you want treated in a specific way, you create a
// `LineHandlerFactory`. `LineHandlerFactory.New()` then creates a new
// `LineHandler` for each specific file. (This mechanism lets parsers be
// stateful, because you end up with one parser instance per file.)
package handlers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
	"github.com/honeycombio/honeycomb-kubernetes-agent/processors"
	"github.com/honeycombio/honeycomb-kubernetes-agent/transmission"
	"github.com/honeycombio/honeycomb-kubernetes-agent/unwrappers"
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
	transmitter   transmission.Transmitter
}

func NewLineHandlerFactoryFromConfig(
	config *config.WatcherConfig,
	unwrapper unwrappers.Unwrapper,
	transmitter transmission.Transmitter,
	extraProcessors ...processors.Processor,
) (*LineHandlerFactoryImpl, error) {
	ret := &LineHandlerFactoryImpl{
		config:      config,
		unwrapper:   unwrapper,
		transmitter: transmitter,
	}
	if config.Dataset == "" {
		return nil, fmt.Errorf("Missing dataset in configuration")
	}
	if config.Parser == nil {
		return nil, fmt.Errorf("No parser specified")
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
		path:       path,
		config:     hf.config,
		parser:     hf.parserFactory.New(),
		processors: hf.processors,
		unwrapper:  hf.unwrapper,
	}
	handler.transmitter = hf.transmitter
	return handler
}

type LineHandlerImpl struct {
	path        string
	config      *config.WatcherConfig
	unwrapper   unwrappers.Unwrapper
	parser      parsers.Parser
	processors  []processors.Processor
	transmitter transmission.Transmitter
}

func (h *LineHandlerImpl) Handle(rawLine string) {
	event, err := h.unwrapper.Unwrap(rawLine, h.parser)
	if err != nil {
		logrus.WithError(err).Debug("Failed to parse line")
		return
	}
	event.Dataset = h.config.Dataset
	event.Path = h.path
	for _, p := range h.processors {
		ret := p.Process(event)
		if !ret {
			logrus.Debug("Dropping line after processing")
			return
		}
	}
	logrus.WithField("parsed", event).Debug("Sending line")
	h.transmitter.Send(event)
}
