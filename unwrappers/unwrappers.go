// Package unwrappers contains support for handling log transport formats.
// How's this different from a parser? You can think of unwrappers as handling
// extra structure that's added by log *transport* formats, e.g., by the Docker
// json-file log driver, or by syslog.
package unwrappers

import (
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/honeycomb-kubernetes-agent/parsers"
)

type Unwrapper interface {
	Unwrap(string, parsers.Parser) (*event.Event, error)
}
