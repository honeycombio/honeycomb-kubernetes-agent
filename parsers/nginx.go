package parsers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/honeycombio/gonx"
)

const defaultLogFormat = `"$remote_addr - $remote_user [$time_local] "$request" $status $bytes_sent "$http_referer" "$http_user_agent"`

type NginxParserFactory struct {
	logFormat string
}

func (pf *NginxParserFactory) Init(options map[string]interface{}) error {
	logFormat, ok := options["log_format"]
	if !ok {
		logFormat = defaultLogFormat
	}
	typedLogFormat, ok := logFormat.(string)
	if !ok {
		return fmt.Errorf("Unexpected type for log_format option")
	}
	pf.logFormat = typedLogFormat
	return nil
}

func (pf *NginxParserFactory) New() Parser {
	return &NginxParser{
		gonxParser: gonx.NewParser(pf.logFormat),
	}
}

type NginxParser struct {
	gonxParser *gonx.Parser
}

// This is basically lifted from honeytail

func (p *NginxParser) Parse(line string) (map[string]interface{}, error) {
	gonxEvent, err := p.gonxParser.ParseString(line)
	if err != nil {
		return nil, err
	}
	return typeifyParsedLine(gonxEvent.Fields), nil
}

// typeifyParsedLine attempts to cast numbers in the event to floats or ints
func typeifyParsedLine(pl map[string]string) map[string]interface{} {
	// try to convert numbers, if possible
	msi := make(map[string]interface{}, len(pl))
	for k, v := range pl {
		switch {
		case strings.Contains(v, "."):
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				msi[k] = f
				continue
			}
		case v == "-":
			// no value, don't set a "-" string
			continue
		default:
			i, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				msi[k] = i
				continue
			}
		}
		msi[k] = v
	}
	return msi
}
