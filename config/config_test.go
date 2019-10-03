package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	testFiles := []struct {
		fileName string
		valid    bool
	}{
		{"basic.yaml", true},
		{"parser_with_options.yaml", true},
		{"unknown_parsers.yaml", false},
	}
	for _, tc := range testFiles {
		path, _ := filepath.Abs(filepath.Join("testdata", tc.fileName))
		_, err := ReadFromFile(path)
		if tc.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestAdditionalFieldsParsing(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("testdata", "basic.yaml"))
	c, err := ReadFromFile(path)
	assert.NoError(t, err)

	assert.Equal(t, map[string]interface{}{"foo": 1, "bar": "2"}, c.AdditionalFields)

	path, _ = filepath.Abs(filepath.Join("testdata", "parser_with_options.yaml"))
	c, err = ReadFromFile(path)
	assert.NoError(t, err)

	assert.Equal(t, map[string]interface{}{"foo": "1", "bar": 2}, c.Watchers[0].Processors[0]["additional_fields"])
}

func TestRegexExpressionsArrayParsing(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("testdata", "parser_with_options.yaml"))
	c, err := ReadFromFile(path)
	assert.NoError(t, err)

	assert.Equal(t, "regex", c.Watchers[1].Parser.Name)
	assert.Equal(t, map[string]interface{}{"expressions": []interface{}{"foo", "bar"}}, c.Watchers[1].Parser.Options)
}
