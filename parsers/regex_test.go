package parsers

import (
	"testing"

	"github.com/honeycombio/honeycomb-kubernetes-agent/config"
	"github.com/stretchr/testify/assert"
)

func TestRegexParser(t *testing.T) {
	cfg := &config.ParserConfig{
		Name: "regex", Options: map[string]interface{}{
			"expressions": []interface{}{
				"(?P<species>[A-z]+) (?P<height>[0-9]{2}[0-9]?)",
				"(?P<city>[A-z ]+),(?P<state>[A-z]{2})",
			},
		},
	}

	pf, err := NewParserFactory(cfg)
	assert.Nil(t, err)
	parser := pf.New()

	tc := []struct {
		line     string
		expected map[string]interface{}
		err      bool
	}{
		{line: "walnut 55", expected: map[string]interface{}{"species": "walnut", "height": "55"}, err: false},
		{line: "douglas 105", expected: map[string]interface{}{"species": "douglas", "height": "105"}, err: false},
		{line: "San Francisco,Ca", expected: map[string]interface{}{"city": "San Francisco", "state": "Ca"}, err: false},
		{line: "South Lake Tahoe,CA", expected: map[string]interface{}{"city": "South Lake Tahoe", "state": "CA"}, err: false},
		{line: "the quick brown fox jumped over the lazy dog", expected: nil, err: true},
	}

	for _, tt := range tc {
		parsed, err := parser.Parse(tt.line)
		assert.Equal(t, parsed, tt.expected)
		if tt.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
