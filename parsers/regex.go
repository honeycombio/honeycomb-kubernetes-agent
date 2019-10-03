package parsers

import (
	"fmt"
	"regexp"
)

type RegexFactory struct {
	expressions []string
}

func (rf *RegexFactory) Init(options map[string]interface{}) error {
	if options == nil {
		return fmt.Errorf("regex parser specified but no options defined")
	}

	expressions, ok := options["expressions"].([]interface{})
	if !ok {
		return fmt.Errorf("regex parser missing patterns option")
	}

	rf.expressions = make([]string, len(expressions))
	for i, s := range expressions {
		expression, ok := s.(string)
		if !ok {
			return fmt.Errorf("expected expression %s to be string", s)
		}
		rf.expressions[i] = expression
	}
	return nil
}

func (rf *RegexFactory) New() Parser {
	re := make([]*extRegexp, len(rf.expressions))
	for i, p := range rf.expressions {
		re[i] = &extRegexp{regexp.MustCompile(p)}
	}
	return &RegexParser{re: re}
}

type RegexParser struct {
	re []*extRegexp
}

func (rp *RegexParser) Parse(line string) (map[string]interface{}, error) {
	var captures map[string]interface{}

	for _, re := range rp.re {
		_, captures = re.FindStringSubmatchIfaceMap(line)
		if captures != nil {
			break
		}
	}

	if captures == nil {
		return nil, fmt.Errorf("Couldn't parse line with any supplied regexes: %s", line)
	}
	return captures, nil
}
