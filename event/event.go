package event

import (
	"fmt"
	"strings"
	"time"
)

type Event struct {
	APIKey     string
	Dataset    string
	Path       string
	SampleRate uint
	Timestamp  time.Time
	Data       map[string]interface{}
	RawMessage string
}

func (e *Event) String() string {
	masked := e.APIKey
	if len(masked) > 4 {
		masked = strings.Repeat("X", len(masked)-4) + masked[len(masked)-4:]
	}
	return fmt.Sprintf("{APIKey:%s Dataset:%s Path:%s SampleRate:%d Timestamp:%v Data:%+v}",
		masked, e.Dataset, e.Path, e.SampleRate, e.Timestamp, e.Data)
}
