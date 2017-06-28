package parsers

type Parser interface {
	Parse(line string) (map[string]interface{}, error)
}

type ParserFactory interface {
	Init(options interface{}) error
	New() Parser
}

func NewParserFactory(parserName string, options interface{}) (ParserFactory, error) {
	var factory ParserFactory
	switch parserName {
	case "json":
		factory = &JSONParserFactory{}
	case "nop":
		factory = &NoOpParserFactory{}
	default:
		factory = &NoOpParserFactory{} // Make this permissive while testing
		// TODO switch back to this:
		//return nil, fmt.Errorf("Unknown parser type %s", parserName)
	}
	err := factory.Init(options)
	if err != nil {
		return nil, err
	}
	return factory, nil
}
