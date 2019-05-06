package transmission

import (
	"strings"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

const truncatedLineMax = 1000

type Transmitter interface {
	Send(*event.Event)
}

type HoneycombTransmitter struct{}

func (ht *HoneycombTransmitter) Send(ev *event.Event) {
	libhoneyEvent := libhoney.NewEvent()
	libhoneyEvent.Metadata = ev.RawMessage
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
		// If an event is too large, include the first part of the raw message
		if resp.Err != nil && strings.HasPrefix(resp.Err.Error(), "event exceeds max event size") {
			rawMessage, ok := resp.Metadata.(string)
			// if this is somehow not a string, handle the error normally below
			if ok {
				msgLen := len(rawMessage)
				truncatedMsgLen := truncatedLineMax
				// This shouldn't happen - if the event is too large for the API, it's at least
				// 1000 bytes, but be defensive here anyway
				if msgLen < truncatedMsgLen {
					truncatedMsgLen = msgLen
				}
				logrus.WithFields(logrus.Fields{
					"error":             resp.Err,
					"status":            resp.StatusCode,
					"responseBody":      string(resp.Body),
					"truncated_raw_msg": rawMessage[:truncatedMsgLen],
					"raw_msg_len":       msgLen,
				}).Error("unable to send event to Honeycomb API due to size")
				continue
			}
		}
		if resp.Err != nil || (resp.StatusCode != 200 && resp.StatusCode != 202) {
			logrus.WithFields(logrus.Fields{
				"error":        resp.Err,
				"status":       resp.StatusCode,
				"responseBody": string(resp.Body),
			}).Error("Failed to send event to Honeycomb")
		}
	}
}

// NullTransmitter does nothing
type NullTransmitter struct{}

func (nt *NullTransmitter) Send(ev *event.Event) {}
