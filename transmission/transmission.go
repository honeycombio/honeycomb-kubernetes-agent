package transmission

import (
	"github.com/Sirupsen/logrus"
	libhoney "github.com/honeycombio/libhoney-go"
)

func Start(writeKey string) error {
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
