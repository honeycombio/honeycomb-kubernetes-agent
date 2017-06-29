package parsers

import "github.com/honeycombio/honeycomb-kubernetes-agent/config"

type Parser interface {
	Parse(line string) (map[string]interface{}, error)
}

type ParserFactory interface {
	Init(options interface{}) error
	New() Parser
}

func NewParserFactory(config *config.ParserConfig) (ParserFactory, error) {
	var factory ParserFactory
	switch config.Name {
	case "json":
		factory = &JSONParserFactory{}
	case "nop":
		factory = &NoOpParserFactory{}
	default:
		factory = &NoOpParserFactory{} // Make this permissive while testing
		// TODO switch back to this:
		//return nil, fmt.Errorf("Unknown parser type %s", parserName)
	}
	err := factory.Init(config.Options)
	if err != nil {
		return nil, err
	}
	return factory, nil
}
