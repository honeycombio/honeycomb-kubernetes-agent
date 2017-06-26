package parsers

type Parser interface {
	Init() error
	Parse(line string) (map[string]interface{}, error)
}
