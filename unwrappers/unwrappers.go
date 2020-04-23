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

type InferUnwrapper struct{
	raw RawLogUnwrapper
	json DockerJSONLogUnwrapper
	cri CriLogUnwrapper
}

func (w *InferUnwrapper) Unwrap(s string, p parsers.Parser) (*event.Event, error) {
	if s[0] == '{' {
		// Scan for the start of a JSON blob.
		return w.json.Unwrap(s, p)
	} else if s[4] == '-' && s[7] == '-' && s[10] == 'T' {
		// Scan for what looks like an RFC3339 timestamp.
		return w.cri.Unwrap(s, p)
	} else {
		// Treat as raw.
		return w.raw.Unwrap(s, p)
	}
}
