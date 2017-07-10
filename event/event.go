package event

import "time"

type Event struct {
	Dataset    string
	SampleRate uint
	Timestamp  time.Time
	Data       map[string]interface{}
}
