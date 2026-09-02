package event

import "time"

type Event struct {
	APIKey     string
	Dataset    string
	Path       string
	SampleRate uint
	Timestamp  time.Time
	Data       map[string]interface{}
	RawMessage string
}
