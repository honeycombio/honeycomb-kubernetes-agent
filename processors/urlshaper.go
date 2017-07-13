package processors

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/urlshaper"
	"github.com/mitchellh/mapstructure"
)

type RequestShaper struct {
	config *requestShaperConfig
	shaper *urlshaper.Parser
}

type requestShaperConfig struct {
	Field     string
	Prefix    string
	Patterns  []string
	QueryKeys []string `yaml:"queryKeys"`
}

func (r *RequestShaper) Init(options map[string]interface{}) error {
	config := &requestShaperConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}
	r.config = config
	r.shaper = &urlshaper.Parser{}
	for _, patternString := range config.Patterns {
		pat := urlshaper.Pattern{Pat: patternString}
		if err := pat.Compile(); err != nil {
			return fmt.Errorf("Invalid pattern %s", patternString)
		}
		r.shaper.Patterns = append(r.shaper.Patterns,
			&pat)
	}

	return nil
}

// This is mostly borrowed from github.com/honeycombio/honeytail
func (r *RequestShaper) Process(ev *event.Event) bool {
	data := ev.Data
	val, ok := data[r.config.Field]
	if !ok {
		return true
	}
	valString, ok := val.(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"key":   r.config.Field,
			"value": val}).
			Debug("Not shaping field of non-string type")
	}

	result := make(map[string]interface{})
	parts := strings.Split(valString, " ")
	var path string
	if len(parts) == 3 {
		result["method"] = parts[0]
		result["protocol_version"] = parts[2]
		path = parts[1]
	} else {
		path = parts[0]
	}

	shapedPath, err := r.shaper.Parse(path)
	if err != nil {
		return true
	}

	result["uri"] = shapedPath.URI
	result["path"] = shapedPath.Path
	result["query"] = shapedPath.Query
	for k, v := range shapedPath.QueryFields {
		for _, whitelistedKey := range r.config.QueryKeys {
			if whitelistedKey == k {
				sort.Strings(v)
				result["query_"+k] = strings.Join(v, ", ")
			}
		}
	}
	for k, v := range shapedPath.PathFields {
		result["path_"+k] = v[0]
	}
	result["shape"] = shapedPath.Shape
	result["pathshape"] = shapedPath.PathShape
	result["queryshape"] = shapedPath.QueryShape

	prefix := r.config.Prefix + r.config.Field
	for k, v := range result {
		data[prefix+"_"+k] = v
	}
	return true
}
