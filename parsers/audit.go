package parsers

import lru "github.com/hashicorp/golang-lru"

var keyvalOptions = map[string]interface{}{"prefixRegex": "(?P<timestamp>[0-9TZ:+.-]+) AUDIT: "}

const cacheSize = 128

type AuditParserFactory struct {
	keyvalParserFactory *KeyvalParserFactory
}

func (pf *AuditParserFactory) Init(options map[string]interface{}) error {
	pf.keyvalParserFactory = &KeyvalParserFactory{}
	pf.keyvalParserFactory.Init(keyvalOptions)
	return nil
}

func (pf *AuditParserFactory) New() Parser {
	cache, _ := lru.New(cacheSize)
	return &AuditParser{
		keyvalParser: pf.keyvalParserFactory.New(),
		cache:        cache,
	}
}

type AuditParser struct {
	keyvalParser Parser
	cache        *lru.Cache
}

func (p *AuditParser) Parse(line string) (map[string]interface{}, error) {
	data, err := p.keyvalParser.Parse(line)
	if err != nil {
		return nil, err
	}
	if id, ok := data["id"]; ok {
		if cached, ok := p.cache.Peek(id); ok {
			if prior, ok := cached.(map[string]interface{}); ok {
				p.cache.Remove(id)
				for k, v := range prior {
					data[k] = v
				}
				return data, nil
			}
		} else {
			p.cache.Add(id, data)
			return nil, nil
		}
	}
	return data, nil
}
