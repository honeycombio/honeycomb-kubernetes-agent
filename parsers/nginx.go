package parsers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/honeycombio/gonx"
)

// nginx's default log format
const defaultLogFormat = `$remote_addr - $remote_user [$time_local] "$request" $status $bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for"`

// envoy's default log format
// https://envoyproxy.github.io/envoy/configuration/http_conn_man/access_log.html#config-http-con-manager-access-log-default-format
const envoyLogFormat = `[$timestamp] "$request" $status_code $response_flags $bytes_received $bytes_sent $duration $x_envoy_upstream_service_time "$x_forwarded_for" "$user_agent" "$x_request_id" "$authority" "$upstream_host"`

// nginx ingress default log format
// https://github.com/kubernetes/ingress-nginx/blob/9c6201b79a8b4/internal/ingress/controller/config/config.go#L53
const nginxIngressLogFormat = `$the_real_ip - [$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`

type NginxParserFactory struct {
	parserName string
	logFormat  string
}

func (pf *NginxParserFactory) Init(options map[string]interface{}) error {
	logFormat := defaultLogFormat
	if pf.parserName == "envoy" {
		logFormat = envoyLogFormat
	}
	if pf.parserName == "nginx-ingress" {
		logFormat = nginxIngressLogFormat
	}
	if logFormatOption, ok := options["log_format"]; ok {
		typedLogFormatOption, ok := logFormatOption.(string)
		if !ok {
			return fmt.Errorf("Unexpected type for log_format option (expected string, got %v", reflect.TypeOf(logFormatOption))
		}

		switch typedLogFormatOption {
		case "default":
			logFormat = defaultLogFormat
		case "envoy":
			logFormat = envoyLogFormat
		case "nginx-ingress":
			logFormat = nginxIngressLogFormat
		default:
			logFormat = typedLogFormatOption
		}
	}

	pf.logFormat = logFormat
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
