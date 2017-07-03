package unwrappers

import "github.com/honeycombio/honeycomb-kubernetes-agent/parsers"

type RawLogUnwrapper struct{}

func (u *RawLogUnwrapper) Unwrap(rawLine string, parser parsers.Parser) (map[string]interface{}, error) {
	return parser.Parse(rawLine)
}
