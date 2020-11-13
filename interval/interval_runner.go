// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package interval

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Runner takes a list of `Runnable`s, calls Setup() on each of them and then
// calls Run() on each of them, using a Ticker, sequentially and within the same
// goroutine. Call Stop() to turn off the Ticker.
type Runner struct {
	name      string
	runnables []Runnable
	ticker    *time.Ticker
	logger    *logrus.Entry
}

// NewRunner creates a new interval runner. Pass in a duration (time between
// calls) and one or more Runnables to be run on the defined interval.
func NewRunner(name string, interval time.Duration, runnables ...Runnable) *Runner {
	return &Runner{
		name:      name,
		runnables: runnables,
		ticker:    time.NewTicker(interval),
		logger:    logrus.WithFields(logrus.Fields{"runner.name": name}),
	}
}

// Runnable must be implemented by types passed into the Runner constructor.
type Runnable interface {
	// called once at Start() time
	Setup() error
	// called on the interval defined by the
	// duration passed into NewRunner
	Run() error
}

// Start kicks off this Runner. Calls Setup() and Run() on the passed-in
// Runnables.
func (r *Runner) Start() error {
	r.logger.Debug("Starting Runner...")
	err := r.setup()
	if err != nil {
		r.logger.WithError(err).Error("Failed to setup Runner")
		return err
	}
	err = r.run()
	if err != nil {
		r.logger.WithError(err).Error("Failed to run Runner")
		return err
	}
	r.logger.Debug("Runner Started")
	return nil
}

func (r *Runner) setup() error {
	for _, runnable := range r.runnables {
		err := runnable.Setup()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) run() error {
	for range r.ticker.C {
		for _, runnable := range r.runnables {
			err := runnable.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop turns off this Runner's ticker.
func (r *Runner) Stop() {
	r.ticker.Stop()
}
