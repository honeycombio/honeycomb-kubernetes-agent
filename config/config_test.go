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
		{"1.yaml", true},
		{"2.yaml", true},
		{"3.yaml", false},
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
