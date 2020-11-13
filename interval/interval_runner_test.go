// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package interval

import (
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	f := &fakeRunnable{}
	s := NewRunner("test", time.Second, f)
	go func() {
		_ = s.Start()
	}()
	s.Stop()
	// getting here is success
}

type fakeRunnable struct {
}

func (t *fakeRunnable) Setup() error {
	return nil
}

func (fakeRunnable) Run() error {
	return nil
}
