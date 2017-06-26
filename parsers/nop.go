package parsers

// Doesn't do any parsing
type NoOpParser struct{}

func (p *NoOpParser) Init() error { return nil }
func (p *NoOpParser) Parse(line string) (map[string]interface{}, error) {
	return map[string]interface{}{"log": line}, nil
}
