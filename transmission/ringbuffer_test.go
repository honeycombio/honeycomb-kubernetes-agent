package transmission

import (
	"testing"
	"time"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	rb := NewRingBuffer(100, 0)

	for i := uint64(0); i < 10; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	ev, ok := rb.Get(5)
	assert.Equal(t, ok, true, "Event not found")
	assert.Equal(t, uint64(5), ev.Data["item"], "Event data incorrect")

	ev, ok = rb.Get(8)
	assert.Equal(t, ok, true, "Event not found")
	assert.Equal(t, uint64(8), ev.Data["item"], "Event data incorrect")

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

	ev, ok := rb.Get(5)
	assert.Equal(t, ok, false, "Event found when it should of been pushed out")

	ev, ok = rb.Get(105)
	assert.Equal(t, ok, true, "Event not found")
	assert.Equal(t, uint64(105), ev.Data["item"], "Event data incorrect")
}

func TestExpire(t *testing.T) {
	rb := NewRingBuffer(100, 1)

	for i := uint64(0); i < 10; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	time.Sleep(2 * time.Second)

	for i := uint64(10); i < 20; i++ {
		ev := &event.Event{
			Data: map[string]interface{}{
				"item": i,
			},
		}
		rb.Add(i, ev)
	}

	ev, ok := rb.Get(5)
	assert.Equal(t, ok, false, "Event found when it should expired")

	ev, ok = rb.Get(15)
	assert.Equal(t, ok, true, "Event not found")
	assert.Equal(t, uint64(15), ev.Data["item"], "Event data incorrect")
}
