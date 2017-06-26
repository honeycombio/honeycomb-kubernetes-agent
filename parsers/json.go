package parsers

import "encoding/json"

// Parses line as JSON
type JSONParser struct{}

func (p *JSONParser) Init() error { return nil }

func (p *JSONParser) Parse(line string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
