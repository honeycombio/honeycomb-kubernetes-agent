package event

import "time"

type Event struct {
	Dataset    string
	Path       string
	SampleRate uint
	Timestamp  time.Time
	Data       map[string]interface{}
}
