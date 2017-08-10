package parsers

// glog is a logging format from Google that's used by Kubernetes components
// (API server, etc.)

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// The only reference I can find for this format is:
// https://github.com/google/glog/blob/master/src/logging.cc#L1077
const lineformat = `(?P<level>[IWEF])(?P<month>[0-9]{2})(?P<day>[0-9]{2}) (?P<hour>[0-9]{2}):(?P<minute>[0-9]{2}):(?P<second>[0-9]{2})\.(?P<microsecond>[0-9]*)\s+(?P<threadid>[0-9]*) (?P<filename>[^:]*):(?P<lineno>[0-9]*)\] (?P<message>.*)`

var levels = map[string]string{
	"I": "info",
	"W": "warning",
	"E": "error",
	"F": "fatal",
}

type GlogParser struct {
	re       *extRegexp
	inFlight map[string]interface{}
}

// TODO: we should support concatenating multiline log statements, because kube
// api server logs, for example, contain lines such as the following (ugh).
// I0720 00:23:31.949027       5 trace.go:61] Trace "GuaranteedUpdate etcd3: *api.Node" (started 2017-07-20 00:23:30.517702742 +0000 UTC):
// [68.006µs] [68.006µs] initial value restored
// [1.03334ms] [965.334µs] Transaction prepared
// [1.431248874s] [1.430215534s] Transaction committed
// "GuaranteedUpdate etcd3: *api.Node" [1.431304615s] [55.741µs] END
func (p *GlogParser) Parse(line string) (map[string]interface{}, error) {
	_, captures := p.re.FindStringSubmatchMap(line)

	if captures == nil {
		return nil, fmt.Errorf("Couldn't parse line as glog line: %s", line)
	}

	ret := make(map[string]interface{}, 0)
	if level, ok := levels[captures["level"]]; ok {
		ret["level"] = level
	} else {
		ret["level"] = captures["level"]
	}
	ret["threadid"] = captures["threadid"]
	ret["filename"] = captures["filename"]
	ret["lineno"] = captures["lineno"]
	ret["message"] = captures["message"]

	ts, err := parseGlogTimestamp(
		captures["month"], captures["day"], captures["hour"], captures["minute"], captures["second"], captures["microsecond"])
	if err == nil {
		ret["glog_timestamp"] = ts
	}

	return ret, nil
}

type GlogParserFactory struct{}

func (pf *GlogParserFactory) Init(options map[string]interface{}) error { return nil }

func (pf *GlogParserFactory) New() Parser {
	return &GlogParser{
		re: &extRegexp{regexp.MustCompile(lineformat)},
	}
}

func parseGlogTimestamp(month string, day string, hour string, minute string, second string, microsecond string) (time.Time, error) {
	year := time.Now().Year()
	var err error
	atoi := func(raw string) int {
		v, newErr := strconv.Atoi(raw)
		err = newErr
		return v
	}
	return time.Date(year, time.Month(atoi(month)), atoi(day), atoi(hour), atoi(minute), atoi(second), atoi(microsecond)*1e3, time.UTC), err
}
