package parsers

// Doesn't do any parsing
type NoOpParser struct{}

func (p *NoOpParser) Parse(line string) (map[string]interface{}, error) {
	return map[string]interface{}{"log": line}, nil
}

type NoOpParserFactory struct{}

func (pf *NoOpParserFactory) Init(options map[string]interface{}) error { return nil }

func (n *NoOpParserFactory) New() Parser {
	return &NoOpParser{}
}
