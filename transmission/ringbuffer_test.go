package transmission

import (
	"testing"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	rb := NewRingBuffer(100, 0)

	assert.True(t, rb.enabled(), "A buffer of size greater than 0 is enabled")

	for i := uint64(0); i < 10; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	ev, ok := rb.Get(5)
	assert.True(t, ok, "Expected event wasn't found in the buffer")
	assert.Equal(t, uint64(5), ev.Data["item"], "Event found didn't match expected data")

	ev, ok = rb.Get(8)
	assert.True(t, ok, "Expected event wasn't found in the buffer")
	assert.Equal(t, uint64(8), ev.Data["item"], "Event found didn't match expected data")
}

func TestGetWhenDisabled(t *testing.T) {
	rb := NewRingBuffer(0, 0)

	assert.False(t, rb.enabled(), "A buffer of size 0 is not enabled")

	for i := uint64(0); i < 10; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	ev, ok := rb.Get(5)
	assert.False(t, ok, "Retrieving an event from a disabled buffer is not ok")
	assert.Nil(t, ev, "A disabled buffer will return nil, not an event")
}

func TestRingOverflow(t *testing.T) {
	rb := NewRingBuffer(100, 0)

	for i := uint64(0); i < 100; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	for i := uint64(100); i < 110; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	_, ok := rb.Get(5)
	assert.False(t, ok, "Found an event on the ring that should have been evicted by the size limit")

	ev, ok := rb.Get(105)
	assert.True(t, ok, "Expected event wasn't found in the buffer")
	assert.Equal(t, uint64(105), ev.Data["item"], "Event found didn't match expected data")
}

func TestExpire(t *testing.T) {
	rb := NewRingBuffer(100, 1*time.Second)

	// this range of events will expire
	for i := uint64(0); i < 10; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	// sleep for more than the TTL
	time.Sleep(2 * time.Second)

	// this range of events will be alive
	for i := uint64(10); i < 20; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	_, ok := rb.Get(5)
	assert.False(t, ok, "Found an event on the ring that should have expired")

	ev, ok := rb.Get(15)
	assert.True(t, ok, "Expected unexpired event wasn't found")
	assert.Equal(t, uint64(15), ev.Data["item"], "Event found didn't match expected data")
}
