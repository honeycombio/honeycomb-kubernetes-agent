package unwrappers

import "github.com/honeycombio/honeycomb-kubernetes-agent/parsers"

type Unwrapper interface {
	Unwrap(string, parsers.Parser) (map[string]interface{}, error)
}
