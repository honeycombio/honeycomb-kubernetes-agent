package parsers

import (
	"fmt"
	"regexp"
	"time"
)

// The only reference I can find for this format is:
// http://build47.com/redis-log-format-levels/
// This formatter targets Redis 3+, which was released in January 2016
const redisLineFormat = `(?P<pid>[0-9]+):(?P<role>[XCSM])\s(?P<timestamp>[0-9]{2}\s[A-Za-z]*\s[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+)\s(?P<level>[.\-*#])\s(?P<message>.*)`
const redisDateFormat = "02 Jan 15:04:05.000"

var redisRoles = map[string]string{
	"X": "sentinel",
	"C": "child",
	"S": "slave",
	"M": "master",
}

var redisLevels = map[string]string{
	".": "debug",
	"-": "verbose",
	"*": "notice",
	"#": "warning",
}

type RedisParser struct {
	re *extRegexp
}

func (p *RedisParser) Parse(line string) (map[string]interface{}, error) {
	_, captures := p.re.FindStringSubmatchMap(line)

	if captures == nil {
		return nil, fmt.Errorf("Couldn't parse line as Redis line: %s", line)
	}

	ret := make(map[string]interface{}, 0)
	if level, ok := redisLevels[captures["level"]]; ok {
		ret["level"] = level
	} else {
		ret["level"] = captures["level"]
	}
	if role, ok := redisRoles[captures["role"]]; ok {
		ret["role"] = role
	} else {
		ret["role"] = captures["role"]
	}
	ret["pid"] = captures["pid"]
	ret["message"] = captures["message"]

	ts, err := time.Parse(redisDateFormat, captures["timestamp"])
	if err == nil {
		ts = ts.AddDate(time.Now().Year(), 0, 0)
		ret["redis_timestamp"] = ts
	}

	return ret, nil
}

type RedisParserFactory struct{}

func (pf *RedisParserFactory) Init(options map[string]interface{}) error { return nil }

func (pf *RedisParserFactory) New() Parser {
	return &RedisParser{
		re: &extRegexp{regexp.MustCompile(redisLineFormat)},
	}
}
