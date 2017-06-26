package processors

type Processor interface {
	Process(data map[string]interface{})
}
