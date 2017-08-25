package parsers

import (
	"fmt"
	"strconv"

	"github.com/kr/logfmt"
)

type KeyvalParserFactory struct {
	prefixExpr      string
	prefixRegex     *extRegexp
	timeFieldName   string
	timeFieldFormat string
}

func (pf *KeyvalParserFactory) Init(options map[string]interface{}) error {
	if prefixExpr, ok := options["prefixRegex"]; ok {
		typedPrefixExpr, ok := prefixExpr.(string)
		if !ok {
			return fmt.Errorf("Invalid type for prefixRegex option (expected string)")
		}
		if typedPrefixExpr != "" {
			re, err := newExtRegexp("^" + typedPrefixExpr) // only match start of line
			if err != nil {
				return fmt.Errorf("Invalid regex value for prefixRegex option: `%s`", typedPrefixExpr)
			}
			pf.prefixRegex = re
			pf.prefixExpr = typedPrefixExpr
		}
	}
	return nil
}

func (pf *KeyvalParserFactory) New() Parser {
	return &KeyvalParser{
		prefixExpr:  pf.prefixExpr,
		prefixRegex: pf.prefixRegex,
	}
}

type KeyvalParser struct {
	prefixExpr  string
	prefixRegex *extRegexp
}

func (p *KeyvalParser) Parse(line string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	var prefixLength int
	if p.prefixRegex != nil {
		prefixMatch, prefixCaptures := p.prefixRegex.FindStringSubmatchMap(line)
		if prefixMatch == "" {
			return nil, fmt.Errorf("Couldn't match line prefix %s for line %s", p.prefixExpr, line)
		}
		for k, v := range prefixCaptures {
			ret[k] = v
		}
		prefixLength = len(prefixMatch)
	}

	err := logfmt.Unmarshal([]byte(line[prefixLength:]), getHandler(ret))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func getHandler(ret map[string]interface{}) logfmt.HandlerFunc {
	return func(key, val []byte) error {
		keyStr := string(key)
		valStr := string(val)
		if b, err := strconv.ParseBool(valStr); err == nil {
			ret[keyStr] = b
			return nil
		}
		if i, err := strconv.Atoi(valStr); err == nil {
			ret[keyStr] = i
			return nil
		}
		if f, err := strconv.ParseFloat(valStr, 64); err == nil {
			ret[keyStr] = f
			return nil
		}
		ret[keyStr] = valStr
		return nil
	}
}
