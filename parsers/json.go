package parsers

import "encoding/json"

// Parses line as JSON
type JSONParser struct{}

func (p *JSONParser) Parse(line string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type JSONParserFactory struct{}

func (pf *JSONParserFactory) Init(options map[string]interface{}) error { return nil }

func (pf *JSONParserFactory) New() Parser {
	return &JSONParser{}
}
