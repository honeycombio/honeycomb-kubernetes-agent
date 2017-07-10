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
	libhoneyEvent.SampleRate = ev.SampleRate
	libhoneyEvent.Timestamp = ev.Timestamp
	libhoneyEvent.Add(ev.Data)
	libhoneyEvent.Send()
}

func InitLibhoney(writeKey string) error {
	err := libhoney.Init(libhoney.Config{WriteKey: writeKey})
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
