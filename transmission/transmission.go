package transmission

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

const truncatedLineMax = 1000

var (
	eventCounter uint64
	ringBuffer   *RingBuffer
)

type Transmitter interface {
	Send(*event.Event)
}

type HoneycombTransmitter struct {
}

func (ht *HoneycombTransmitter) Send(ev *event.Event) {

	libhoneyEvent := libhoney.NewEvent()
	libhoneyEvent.Dataset = ev.Dataset
	if ev.SampleRate != 0 {
		libhoneyEvent.SampleRate = ev.SampleRate
	}
	libhoneyEvent.Timestamp = ev.Timestamp

	err := libhoneyEvent.Add(ev.Data)
	if err != nil {
		// event isn't a proper type that can be added
		logrus.WithFields(logrus.Fields{
			"error":      err.Error(),
			"rawMessage": ev.RawMessage,
		}).Error("Unable to create libhoney event.")
		return
	}

	// generate auto-incrementing value to be used as buffer key
	key := atomic.AddUint64(&eventCounter, 1)
	if key == 0 {
		// if the counter is rolled back to 0, add 1 more
		key = atomic.AddUint64(&eventCounter, 1)
	}
	libhoneyEvent.Metadata = key
	ringBuffer.Add(key, ev)

	// send event
	err = libhoneyEvent.SendPresampled()
	if err != nil {
		// error while sending
		logrus.WithFields(logrus.Fields{
			"error":      err.Error(),
			"rawMessage": ev.RawMessage,
		}).Error("Unable to send libhoney event.")
		return
	}
}

func InitLibhoney(apiKey string, apiHost string, bufferSize int, bufferExpire time.Duration) error {
	err := libhoney.Init(libhoney.Config{
		APIKey:      apiKey,
		APIHost:     apiHost,
		BlockOnSend: true,
	})
	if err != nil {
		return err
	}

	ringBuffer = NewRingBuffer(bufferSize, bufferExpire)

	go readResponses()
	return nil
}

func readResponses() {
	for resp := range libhoney.TxResponses() {
		if resp.Err != nil && strings.HasPrefix(resp.Err.Error(), "event exceeds max event size") {
			// error sending event due to size, try to find it in cache
			key, ok := resp.Metadata.(uint64)
			if ok {
				ev, exists := ringBuffer.Get(key)

				if exists {
					rawMessage := ev.RawMessage

					// include the first part of the raw message
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
					}).Error("Unable to send event to Honeycomb API due to size")

					continue
				}
			}

			// if we get here we couldn't find the response metadata or event in send buffer
			logrus.WithFields(logrus.Fields{
				"error":        resp.Err,
				"status":       resp.StatusCode,
				"responseBody": string(resp.Body),
			}).Error("Unable to send event to Honeycomb API due to size. Event not found in local send buffer.")

		} else if resp.Err != nil || (resp.StatusCode != 200 && resp.StatusCode != 202) {
			// error sending event, try to find it in buffer
			key, ok := resp.Metadata.(uint64)
			if ok {
				ev, exists := ringBuffer.Get(key)

				if exists {
					// event found, we can retry it
					logrus.WithFields(logrus.Fields{
						"error":        resp.Err,
						"status":       resp.StatusCode,
						"responseBody": string(resp.Body),
					}).Debug("Failed to send event to Honeycomb. Will retry.")

					// Resend event
					ht := &HoneycombTransmitter{}
					ht.Send(ev)

					continue
				}
			}

			// if we get here we couldn't find the response metadata or event in send buffer
			// retry isn't possible
			logrus.WithFields(logrus.Fields{
				"error":        resp.Err,
				"status":       resp.StatusCode,
				"responseBody": string(resp.Body),
			}).Error("Failed to send event to Honeycomb. Unable to retry.")
		}
	}
}

// NullTransmitter does nothing
type NullTransmitter struct{}

func (nt *NullTransmitter) Send(_ *event.Event) {}
