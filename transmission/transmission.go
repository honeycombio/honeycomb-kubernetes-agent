package transmission

import (
	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	libhoney "github.com/honeycombio/libhoney-go"
)

type Transmitter interface {
	Send(*event.Event)
}

type HoneycombTransmitter struct{}

func (ht *HoneycombTransmitter) Send(ev *event.Event) {
	libhoneyEvent := libhoney.NewEvent()
	libhoneyEvent.Dataset = ev.Dataset
	if ev.SampleRate != 0 {
		libhoneyEvent.SampleRate = ev.SampleRate
	}
	libhoneyEvent.Timestamp = ev.Timestamp
	libhoneyEvent.Add(ev.Data)
	libhoneyEvent.SendPresampled()
}

func InitLibhoney(writeKey string, apiHost string) error {
	err := libhoney.Init(libhoney.Config{
		WriteKey: writeKey,
		APIHost:  apiHost,
	})
	if err != nil {
		return err
	}
	go readResponses()
	return nil
}

// TODO: support retrying, etc.
func readResponses() {
	for resp := range libhoney.Responses() {
		if resp.Err != nil || (resp.StatusCode != 200 && resp.StatusCode != 202) {
			logrus.WithFields(logrus.Fields{
				"error":        resp.Err,
				"status":       resp.StatusCode,
				"responseBody": string(resp.Body),
			}).Error("Failed to send event to Honeycomb")
		}
	}
}
